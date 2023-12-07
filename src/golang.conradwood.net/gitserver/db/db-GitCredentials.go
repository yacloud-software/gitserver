package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBGitCredentials
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence gitcredentials_seq;

Main Table:

 CREATE TABLE gitcredentials (id integer primary key default nextval('gitcredentials_seq'),userid text not null  ,host text not null  ,path text not null  ,username text not null  ,password text not null  ,expiry integer not null  );

Alter statements:
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS host text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS path text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS username text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS password text not null default '';
ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS expiry integer not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE gitcredentials_archive (id integer unique not null,userid text not null,host text not null,path text not null,username text not null,password text not null,expiry integer not null);
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
	default_def_DBGitCredentials *DBGitCredentials
)

type DBGitCredentials struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBGitCredentials() *DBGitCredentials {
	if default_def_DBGitCredentials != nil {
		return default_def_DBGitCredentials
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBGitCredentials(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBGitCredentials = res
	return res
}
func NewDBGitCredentials(db *sql.DB) *DBGitCredentials {
	foo := DBGitCredentials{DB: db}
	foo.SQLTablename = "gitcredentials"
	foo.SQLArchivetablename = "gitcredentials_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBGitCredentials) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBGitCredentials", "insert into "+a.SQLArchivetablename+" (id,userid, host, path, username, password, expiry) values ($1,$2, $3, $4, $5, $6, $7) ", p.ID, p.UserID, p.Host, p.Path, p.Username, p.Password, p.Expiry)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBGitCredentials) Save(ctx context.Context, p *savepb.GitCredentials) (uint64, error) {
	qn := "DBGitCredentials_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (userid, host, path, username, password, expiry) values ($1, $2, $3, $4, $5, $6) returning id", p.UserID, p.Host, p.Path, p.Username, p.Password, p.Expiry)
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
func (a *DBGitCredentials) SaveWithID(ctx context.Context, p *savepb.GitCredentials) error {
	qn := "insert_DBGitCredentials"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,userid, host, path, username, password, expiry) values ($1,$2, $3, $4, $5, $6, $7) ", p.ID, p.UserID, p.Host, p.Path, p.Username, p.Password, p.Expiry)
	return a.Error(ctx, qn, e)
}

func (a *DBGitCredentials) Update(ctx context.Context, p *savepb.GitCredentials) error {
	qn := "DBGitCredentials_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set userid=$1, host=$2, path=$3, username=$4, password=$5, expiry=$6 where id = $7", p.UserID, p.Host, p.Path, p.Username, p.Password, p.Expiry, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBGitCredentials) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBGitCredentials_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBGitCredentials) ByID(ctx context.Context, p uint64) (*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No GitCredentials with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) GitCredentials with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBGitCredentials) TryByID(ctx context.Context, p uint64) (*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) GitCredentials with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBGitCredentials) All(ctx context.Context) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" order by id")
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

// get all "DBGitCredentials" rows with matching UserID
func (a *DBGitCredentials) ByUserID(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where userid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikeUserID(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where userid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Host
func (a *DBGitCredentials) ByHost(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where host = $1", p)
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
func (a *DBGitCredentials) ByLikeHost(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikeHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where host ilike $1", p)
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

// get all "DBGitCredentials" rows with matching Path
func (a *DBGitCredentials) ByPath(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByPath"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where path = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPath: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPath: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikePath(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikePath"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where path ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPath: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Username
func (a *DBGitCredentials) ByUsername(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByUsername"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where username = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUsername: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUsername: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikeUsername(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikeUsername"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where username ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUsername: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUsername: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Password
func (a *DBGitCredentials) ByPassword(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByPassword"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where password = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPassword: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPassword: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGitCredentials) ByLikePassword(ctx context.Context, p string) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikePassword"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where password ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPassword: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByPassword: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGitCredentials" rows with matching Expiry
func (a *DBGitCredentials) ByExpiry(ctx context.Context, p uint32) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByExpiry"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where expiry = $1", p)
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
func (a *DBGitCredentials) ByLikeExpiry(ctx context.Context, p uint32) ([]*savepb.GitCredentials, error) {
	qn := "DBGitCredentials_ByLikeExpiry"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,userid, host, path, username, password, expiry from "+a.SQLTablename+" where expiry ilike $1", p)
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
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBGitCredentials) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.GitCredentials, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBGitCredentials) Tablename() string {
	return a.SQLTablename
}

func (a *DBGitCredentials) SelectCols() string {
	return "id,userid, host, path, username, password, expiry"
}
func (a *DBGitCredentials) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".userid, " + a.SQLTablename + ".host, " + a.SQLTablename + ".path, " + a.SQLTablename + ".username, " + a.SQLTablename + ".password, " + a.SQLTablename + ".expiry"
}

func (a *DBGitCredentials) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.GitCredentials, error) {
	var res []*savepb.GitCredentials
	for rows.Next() {
		foo := savepb.GitCredentials{}
		err := rows.Scan(&foo.ID, &foo.UserID, &foo.Host, &foo.Path, &foo.Username, &foo.Password, &foo.Expiry)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}

/**********************************************************************
* Helper to create table and columns
**********************************************************************/
func (a *DBGitCredentials) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid text not null ,host text not null ,path text not null ,username text not null ,password text not null ,expiry integer not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),userid text not null ,host text not null ,path text not null ,username text not null ,password text not null ,expiry integer not null );`,
		`ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS host text not null default '';`,
		`ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS path text not null default '';`,
		`ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS username text not null default '';`,
		`ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS password text not null default '';`,
		`ALTER TABLE gitcredentials ADD COLUMN IF NOT EXISTS expiry integer not null default 0;`,

		`ALTER TABLE gitcredentials_archive ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE gitcredentials_archive ADD COLUMN IF NOT EXISTS host text not null default '';`,
		`ALTER TABLE gitcredentials_archive ADD COLUMN IF NOT EXISTS path text not null default '';`,
		`ALTER TABLE gitcredentials_archive ADD COLUMN IF NOT EXISTS username text not null default '';`,
		`ALTER TABLE gitcredentials_archive ADD COLUMN IF NOT EXISTS password text not null default '';`,
		`ALTER TABLE gitcredentials_archive ADD COLUMN IF NOT EXISTS expiry integer not null default 0;`,
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
func (a *DBGitCredentials) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

