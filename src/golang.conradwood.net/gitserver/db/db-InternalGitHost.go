package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBInternalGitHost
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence internalgithost_seq;

Main Table:

 CREATE TABLE internalgithost (id integer primary key default nextval('internalgithost_seq'),host text not null  ,expiry integer not null  );

Alter statements:
ALTER TABLE internalgithost ADD COLUMN IF NOT EXISTS host text not null default '';
ALTER TABLE internalgithost ADD COLUMN IF NOT EXISTS expiry integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE internalgithost_archive (id integer unique not null,host text not null,expiry integer not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/sql"
	"os"
)

var (
	default_def_DBInternalGitHost *DBInternalGitHost
)

type DBInternalGitHost struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBInternalGitHost() *DBInternalGitHost {
	if default_def_DBInternalGitHost != nil {
		return default_def_DBInternalGitHost
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBInternalGitHost(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBInternalGitHost = res
	return res
}
func NewDBInternalGitHost(db *sql.DB) *DBInternalGitHost {
	foo := DBInternalGitHost{DB: db}
	foo.SQLTablename = "internalgithost"
	foo.SQLArchivetablename = "internalgithost_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBInternalGitHost) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBInternalGitHost", "insert into "+a.SQLArchivetablename+" (id,host, expiry) values ($1,$2, $3) ", p.ID, p.Host, p.Expiry)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBInternalGitHost) Save(ctx context.Context, p *savepb.InternalGitHost) (uint64, error) {
	qn := "DBInternalGitHost_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (host, expiry) values ($1, $2) returning id", a.get_Host(p), a.get_Expiry(p))
	if e != nil {
		return 0, a.Error(ctx, qn, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, qn, fmt.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, qn, fmt.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// Save using the ID specified
func (a *DBInternalGitHost) SaveWithID(ctx context.Context, p *savepb.InternalGitHost) error {
	qn := "insert_DBInternalGitHost"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,host, expiry) values ($1,$2, $3) ", p.ID, p.Host, p.Expiry)
	return a.Error(ctx, qn, e)
}

func (a *DBInternalGitHost) Update(ctx context.Context, p *savepb.InternalGitHost) error {
	qn := "DBInternalGitHost_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set host=$1, expiry=$2 where id = $3", a.get_Host(p), a.get_Expiry(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBInternalGitHost) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBInternalGitHost_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBInternalGitHost) ByID(ctx context.Context, p uint64) (*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,host, expiry from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No InternalGitHost with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) InternalGitHost with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBInternalGitHost) TryByID(ctx context.Context, p uint64) (*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,host, expiry from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) InternalGitHost with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBInternalGitHost) All(ctx context.Context) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,host, expiry from "+a.SQLTablename+" order by id")
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("All: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, fmt.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBInternalGitHost" rows with matching Host
func (a *DBInternalGitHost) ByHost(ctx context.Context, p string) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,host, expiry from "+a.SQLTablename+" where host = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByHost: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBInternalGitHost) ByLikeHost(ctx context.Context, p string) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByLikeHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,host, expiry from "+a.SQLTablename+" where host ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByHost: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByHost: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBInternalGitHost" rows with matching Expiry
func (a *DBInternalGitHost) ByExpiry(ctx context.Context, p uint32) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByExpiry"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,host, expiry from "+a.SQLTablename+" where expiry = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByExpiry: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBInternalGitHost) ByLikeExpiry(ctx context.Context, p uint32) ([]*savepb.InternalGitHost, error) {
	qn := "DBInternalGitHost_ByLikeExpiry"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,host, expiry from "+a.SQLTablename+" where expiry ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByExpiry: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByExpiry: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

func (a *DBInternalGitHost) get_ID(p *savepb.InternalGitHost) uint64 {
	return p.ID
}

func (a *DBInternalGitHost) get_Host(p *savepb.InternalGitHost) string {
	return p.Host
}

func (a *DBInternalGitHost) get_Expiry(p *savepb.InternalGitHost) uint32 {
	return p.Expiry
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBInternalGitHost) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.InternalGitHost, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBInternalGitHost) Tablename() string {
	return a.SQLTablename
}

func (a *DBInternalGitHost) SelectCols() string {
	return "id,host, expiry"
}
func (a *DBInternalGitHost) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".host, " + a.SQLTablename + ".expiry"
}

func (a *DBInternalGitHost) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.InternalGitHost, error) {
	var res []*savepb.InternalGitHost
	for rows.Next() {
		foo := savepb.InternalGitHost{}
		err := rows.Scan(&foo.ID, &foo.Host, &foo.Expiry)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBInternalGitHost) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.InternalGitHost, error) {
	var res []*savepb.InternalGitHost
	for rows.Next() {
		// SCANNER:
		foo := &savepb.InternalGitHost{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.Host
		scanTarget_2 := &foo.Expiry
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2)
		// END SCANNER

		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBInternalGitHost) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),host text not null ,expiry integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),host text not null ,expiry integer not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS host text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS expiry integer not null default 0;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS host text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS expiry integer not null  default 0;`,
	}

	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
	}

	// these are optional, expected to fail
	csql = []string{
		// Indices:

		// Foreign keys:

	}
	for i, c := range csql {
		a.DB.ExecContextQuiet(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBInternalGitHost) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

