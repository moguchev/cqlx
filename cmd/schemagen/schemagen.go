// Copyright (C) 2017 ScyllaDB
// Use of this source code is governed by a ALv2-style
// license that can be found in the LICENSE file.
//
// Modifications Copyright (C) 2026 Leonid Chenskii. Apache-2.0.

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	gocqlv2 "github.com/apache/cassandra-gocql-driver/v2"

	"github.com/moguchev/cqlx"
	_ "github.com/moguchev/cqlx/table"
)

var defaultClusterConfig = gocqlv2.NewCluster()

var (
	defaultQueryTimeout      = defaultClusterConfig.Timeout
	defaultConnectionTimeout = defaultClusterConfig.ConnectTimeout
)

var (
	cmd                           = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagCluster                   = cmd.String("cluster", "127.0.0.1", "a comma-separated list of host:port tuples")
	flagKeyspace                  = cmd.String("keyspace", "", "keyspace to inspect")
	flagPkgname                   = cmd.String("pkgname", "models", "the name you wish to assign to your generated package")
	flagOutput                    = cmd.String("output", "models", "the name of the folder to output to")
	flagOutputDirPerm             = cmd.Uint64("output-dir-perm", 0o755, "output directory permissions")
	flagOutputFilePerm            = cmd.Uint64("output-file-perm", 0o644, "output file permissions")
	flagUser                      = cmd.String("user", "", "user for password authentication")
	flagPassword                  = cmd.String("password", "", "password for password authentication")
	flagIgnoreNames               = cmd.String("ignore-names", "", "a comma-separated list of table or view names to ignore")
	flagQueryTimeout              = cmd.Duration("query-timeout", defaultQueryTimeout, "query timeout ( in seconds )")
	flagConnectionTimeout         = cmd.Duration("connection-timeout", defaultConnectionTimeout, "connection timeout ( in seconds )")
	flagSSLEnableHostVerification = cmd.Bool("ssl-enable-host-verification", false, "don't check server ssl certificate")
	flagSSLCAPath                 = cmd.String("ssl-ca-path", "", "path to ssl CA certificates")
	flagSSLCertPath               = cmd.String("ssl-cert-path", "", "path to ssl certificate")
	flagSSLKeyPath                = cmd.String("ssl-key-path", "", "path to ssl key")
)

//go:embed keyspace.tmpl
var keyspaceTmpl string

func main() {
	err := cmd.Parse(os.Args[1:])
	if err != nil {
		log.Fatalln("can't parse flags")
	}

	if *flagKeyspace == "" {
		log.Fatalln("missing required flag: keyspace")
	}

	if err := schemagen(); err != nil {
		log.Fatalf("failed to generate schema: %s", err)
	}
}

func schemagen() error {
	if err := os.MkdirAll(*flagOutput, os.FileMode(*flagOutputDirPerm)); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	session, err := createSession()
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}
	metadata, err := session.KeyspaceMetadata(*flagKeyspace)
	if err != nil {
		return fmt.Errorf("fetch keyspace metadata: %w", err)
	}
	b, err := renderTemplate(metadata)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	outputPath := path.Join(*flagOutput, *flagPkgname+".go")

	return os.WriteFile(outputPath, b, fs.FileMode(*flagOutputFilePerm))
}

func renderTemplate(md *gocqlv2.KeyspaceMetadata) ([]byte, error) {
	t, err := template.
		New("keyspace.tmpl").
		Funcs(template.FuncMap{"camelize": camelize}).
		Funcs(template.FuncMap{"mapCqlTypeToGoType": mapCqlTypeToGoType}).
		Parse(keyspaceTmpl)
	if err != nil {
		log.Fatalln("unable to parse models template:", err)
	}

	// Remove all tables and views whose names match the filter.
	ignoredNames := make(map[string]struct{})
	for _, ignoredName := range strings.Split(*flagIgnoreNames, ",") {
		ignoredNames[ignoredName] = struct{}{}
	}
	for name := range ignoredNames {
		delete(md.Tables, name)
		delete(md.MaterializedViews, name)
	}

	// The Apache v2 driver does not pre-register UDTs in its type lookup,
	// so KeyspaceMetadata returns columns/UDT field types referencing UDTs
	// as opaque unknownTypeInfo (validator "frozen<udt>") instead of
	// UDTTypeInfo. Rewrite them in-place so the rest of the generator
	// (orphan detection, import collection, mapCqlTypeToGoType) just works.
	resolveUDTReferences(md)

	// Delete a user-defined type (UDT) if it is not used by any column.
	orphanedTypes := make(map[string]struct{})
	for userTypeName := range md.UserTypes {
		if !usedInTables(userTypeName, md.Tables) &&
			!usedInViews(userTypeName, md.MaterializedViews) {
			orphanedTypes[userTypeName] = struct{}{}
		}
	}
	for typeName := range orphanedTypes {
		delete(md.UserTypes, typeName)
	}

	imports := make([]string, 0)
	if len(md.UserTypes) != 0 {
		imports = append(imports, "github.com/moguchev/cqlx")
	}

	updateImports := func(columns map[string]*gocqlv2.ColumnMetadata) {
		for _, c := range columns {
			collectImports(c.Type, &imports)
		}
	}

	// Ensure that for each table and view
	//
	// 1. ordered columns are sorted alphabetically;
	// 2. imports are resolved for column types.
	for _, t := range md.Tables {
		sort.Strings(t.OrderedColumns)
		updateImports(t.Columns)
	}
	for _, v := range md.MaterializedViews {
		if v.BaseTable != nil {
			sort.Strings(v.BaseTable.OrderedColumns)
			updateImports(v.BaseTable.Columns)
		}
	}
	for _, ut := range md.UserTypes {
		for _, ft := range ut.FieldTypes {
			collectImports(ft, &imports)
		}
	}

	buf := &bytes.Buffer{}
	data := map[string]any{
		"PackageName": *flagPkgname,
		"Tables":      md.Tables,
		"Views":       md.MaterializedViews,
		"UserTypes":   md.UserTypes,
		"Imports":     imports,
	}

	if err = t.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("template: %w", err)
	}
	return format.Source(buf.Bytes())
}

// resolveUDTReferences rewrites columns and UDT field types that the Apache v2
// driver returned as opaque unknownTypeInfo into proper gocqlv2.UDTTypeInfo
// values whenever the underlying name matches a UDT in md.UserTypes. This
// works around the driver not pre-registering UDTs in the type parser, which
// otherwise leaves `frozen<udt_name>` columns typed as TypeCustom strings.
func resolveUDTReferences(md *gocqlv2.KeyspaceMetadata) {
	udts := make(map[string]struct{}, len(md.UserTypes))
	for name := range md.UserTypes {
		udts[name] = struct{}{}
	}
	if len(udts) == 0 {
		return
	}

	for _, t := range md.Tables {
		for _, c := range t.Columns {
			c.Type = rewriteUDTTypeInfo(c.Type, udts, md.Name)
		}
	}
	for _, v := range md.MaterializedViews {
		if v.BaseTable == nil {
			continue
		}
		for _, c := range v.BaseTable.Columns {
			c.Type = rewriteUDTTypeInfo(c.Type, udts, md.Name)
		}
	}
	for _, ut := range md.UserTypes {
		for i, ft := range ut.FieldTypes {
			ut.FieldTypes[i] = rewriteUDTTypeInfo(ft, udts, md.Name)
		}
	}
}

// rewriteUDTTypeInfo walks a TypeInfo and replaces any unknownTypeInfo whose
// validator string resolves (after stripping frozen<…> wrappers) to a UDT in
// the provided set. Collection element/key and tuple element types are walked
// recursively.
func rewriteUDTTypeInfo(t gocqlv2.TypeInfo, udts map[string]struct{}, keyspace string) gocqlv2.TypeInfo {
	if t == nil {
		return nil
	}
	switch t.Type() {
	case gocqlv2.TypeList, gocqlv2.TypeSet:
		if ct, ok := t.(gocqlv2.CollectionType); ok {
			ct.Elem = rewriteUDTTypeInfo(ct.Elem, udts, keyspace)
			return ct
		}
	case gocqlv2.TypeMap:
		if ct, ok := t.(gocqlv2.CollectionType); ok {
			ct.Key = rewriteUDTTypeInfo(ct.Key, udts, keyspace)
			ct.Elem = rewriteUDTTypeInfo(ct.Elem, udts, keyspace)
			return ct
		}
	case gocqlv2.TypeTuple:
		if tt, ok := t.(gocqlv2.TupleTypeInfo); ok {
			for i, e := range tt.Elems {
				tt.Elems[i] = rewriteUDTTypeInfo(e, udts, keyspace)
			}
			return tt
		}
	case gocqlv2.TypeCustom:
		// unknownTypeInfo is `type unknownTypeInfo string`; the value's
		// string form is the original validator (e.g. "frozen<album>").
		name := strings.TrimSpace(fmt.Sprintf("%s", t))
		for strings.HasPrefix(name, "frozen<") && strings.HasSuffix(name, ">") {
			name = strings.TrimSpace(name[len("frozen<") : len(name)-1])
		}
		if _, ok := udts[name]; ok {
			return gocqlv2.UDTTypeInfo{Keyspace: keyspace, Name: name}
		}
	}
	return t
}

// collectImports walks a TypeInfo recursively and adds package imports
// (time, gopkg.in/inf.v0, apache gocql) when required by the underlying CQL type.
func collectImports(t gocqlv2.TypeInfo, imports *[]string) {
	if t == nil {
		return
	}
	switch t.Type() {
	case gocqlv2.TypeTimestamp, gocqlv2.TypeDate, gocqlv2.TypeTime:
		appendUnique(imports, "time")
	case gocqlv2.TypeDecimal:
		appendUnique(imports, "gopkg.in/inf.v0")
	case gocqlv2.TypeDuration:
		appendUnique(imports, "github.com/apache/cassandra-gocql-driver/v2")
	case gocqlv2.TypeList, gocqlv2.TypeSet:
		if ct, ok := t.(gocqlv2.CollectionType); ok {
			collectImports(ct.Elem, imports)
		}
	case gocqlv2.TypeMap:
		if ct, ok := t.(gocqlv2.CollectionType); ok {
			collectImports(ct.Key, imports)
			collectImports(ct.Elem, imports)
		}
	case gocqlv2.TypeTuple:
		if tt, ok := t.(gocqlv2.TupleTypeInfo); ok {
			for _, e := range tt.Elems {
				collectImports(e, imports)
			}
		}
	}
}

func appendUnique(s *[]string, v string) {
	if !existsInSlice(*s, v) {
		*s = append(*s, v)
	}
}

func createSession() (cqlx.Session, error) {
	cluster := gocqlv2.NewCluster(clusterHosts()...)

	if *flagUser != "" {
		cluster.Authenticator = gocqlv2.PasswordAuthenticator{
			Username: *flagUser,
			Password: *flagPassword,
		}
	}

	if *flagQueryTimeout >= 0 {
		cluster.Timeout = *flagQueryTimeout
	}
	if *flagConnectionTimeout >= 0 {
		cluster.ConnectTimeout = *flagConnectionTimeout
	}

	if *flagSSLCAPath != "" || *flagSSLCertPath != "" || *flagSSLKeyPath != "" {
		cluster.SslOpts = &gocqlv2.SslOptions{
			EnableHostVerification: *flagSSLEnableHostVerification,
			CaPath:                 *flagSSLCAPath,
			CertPath:               *flagSSLCertPath,
			KeyPath:                *flagSSLKeyPath,
		}
	}

	return cqlx.WrapSession(cluster.CreateSession())
}

func clusterHosts() []string {
	return strings.Split(*flagCluster, ",")
}

func existsInSlice(s []string, v string) bool {
	for _, i := range s {
		if v == i {
			return true
		}
	}

	return false
}

// usedInColumns tests whether the typeName is used in any of the columns,
// walking the TypeInfo recursively (handles UDT, list/set/map, tuple).
func usedInColumns(typeName string, columns map[string]*gocqlv2.ColumnMetadata) bool {
	for _, column := range columns {
		if typeUsesName(column.Type, typeName) {
			return true
		}
	}
	return false
}

// typeUsesName reports whether t refers to a UDT named typeName,
// recursing into collections, maps and tuples.
func typeUsesName(t gocqlv2.TypeInfo, typeName string) bool {
	if t == nil {
		return false
	}
	switch t.Type() {
	case gocqlv2.TypeUDT:
		if udt, ok := t.(gocqlv2.UDTTypeInfo); ok && udt.Name == typeName {
			return true
		}
	case gocqlv2.TypeList, gocqlv2.TypeSet:
		if ct, ok := t.(gocqlv2.CollectionType); ok {
			return typeUsesName(ct.Elem, typeName)
		}
	case gocqlv2.TypeMap:
		if ct, ok := t.(gocqlv2.CollectionType); ok {
			return typeUsesName(ct.Key, typeName) || typeUsesName(ct.Elem, typeName)
		}
	case gocqlv2.TypeTuple:
		if tt, ok := t.(gocqlv2.TupleTypeInfo); ok {
			for _, e := range tt.Elems {
				if typeUsesName(e, typeName) {
					return true
				}
			}
		}
	}
	return false
}

// usedInTables tests whether the typeName is used in any of the columns of
// the provided tables.
func usedInTables(typeName string, tables map[string]*gocqlv2.TableMetadata) bool {
	for _, table := range tables {
		if usedInColumns(typeName, table.Columns) {
			return true
		}
	}
	return false
}

// usedInViews tests whether the typeName is used in any of the columns of
// the provided materialized views (via the underlying base table).
func usedInViews(typeName string, views map[string]*gocqlv2.MaterializedViewMetadata) bool {
	for _, v := range views {
		if v.BaseTable == nil {
			continue
		}
		if usedInColumns(typeName, v.BaseTable.Columns) {
			return true
		}
	}
	return false
}
