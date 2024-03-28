package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBUserRepositoryAccess
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence userrepositoryaccess_seq;

Main Table:

 CREATE TABLE userrepositoryaccess (id integer primary key default nextval('userrepositoryaccess_seq'),repoid bigint not null  ,userid text not null  ,read boolean not null  ,write boolean not null  );

Alter statements:
ALTER TABLE userrepositoryaccess ADD COLUMN IF NOT EXISTS repoid bigint not null default 0;
ALTER TABLE userrepositoryaccess ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE userrepositoryaccess ADD COLUMN IF NOT EXISTS read boolean not null default false;
ALTER TABLE userrepositoryaccess ADD COLUMN IF NOT EXISTS write boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE userrepositoryaccess_archive (id integer unique not null,repoid bigint not null,userid text not null,read boolean not null,write boolean not null);
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
	default_def_DBUserRepositoryAccess *DBUserRepositoryAccess
)

type DBUserRepositoryAccess struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBUserRepositoryAccess() *DBUserRepositoryAccess {
	if default_def_DBUserRepositoryAccess != nil {
		return default_def_DBUserRepositoryAccess
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBUserRepositoryAccess(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBUserRepositoryAccess = res
	return res
}
func NewDBUserRepositoryAccess(db *sql.DB) *DBUserRepositoryAccess {
	foo := DBUserRepositoryAccess{DB: db}
	foo.SQLTablename = "userrepositoryaccess"
	foo.SQLArchivetablename = "userrepositoryaccess_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBUserRepositoryAccess) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBUserRepositoryAccess", "insert into "+a.SQLArchivetablename+" (id,repoid, userid, read, write) values ($1,$2, $3, $4, $5) ", p.ID, p.RepoID, p.UserID, p.Read, p.Write)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBUserRepositoryAccess) Save(ctx context.Context, p *savepb.UserRepositoryAccess) (uint64, error) {
	qn := "DBUserRepositoryAccess_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (repoid, userid, read, write) values ($1, $2, $3, $4) returning id", a.get_RepoID(p), a.get_UserID(p), a.get_Read(p), a.get_Write(p))
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
func (a *DBUserRepositoryAccess) SaveWithID(ctx context.Context, p *savepb.UserRepositoryAccess) error {
	qn := "insert_DBUserRepositoryAccess"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,repoid, userid, read, write) values ($1,$2, $3, $4, $5) ", p.ID, p.RepoID, p.UserID, p.Read, p.Write)
	return a.Error(ctx, qn, e)
}

func (a *DBUserRepositoryAccess) Update(ctx context.Context, p *savepb.UserRepositoryAccess) error {
	qn := "DBUserRepositoryAccess_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repoid=$1, userid=$2, read=$3, write=$4 where id = $5", a.get_RepoID(p), a.get_UserID(p), a.get_Read(p), a.get_Write(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBUserRepositoryAccess) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBUserRepositoryAccess_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBUserRepositoryAccess) ByID(ctx context.Context, p uint64) (*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No UserRepositoryAccess with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) UserRepositoryAccess with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBUserRepositoryAccess) TryByID(ctx context.Context, p uint64) (*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) UserRepositoryAccess with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBUserRepositoryAccess) All(ctx context.Context) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" order by id")
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

// get all "DBUserRepositoryAccess" rows with matching RepoID
func (a *DBUserRepositoryAccess) ByRepoID(ctx context.Context, p uint64) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByRepoID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where repoid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserRepositoryAccess) ByLikeRepoID(ctx context.Context, p uint64) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByLikeRepoID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where repoid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with matching UserID
func (a *DBUserRepositoryAccess) ByUserID(ctx context.Context, p string) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where userid = $1", p)
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
func (a *DBUserRepositoryAccess) ByLikeUserID(ctx context.Context, p string) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where userid ilike $1", p)
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

// get all "DBUserRepositoryAccess" rows with matching Read
func (a *DBUserRepositoryAccess) ByRead(ctx context.Context, p bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByRead"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where read = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRead: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRead: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserRepositoryAccess) ByLikeRead(ctx context.Context, p bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByLikeRead"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where read ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRead: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRead: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBUserRepositoryAccess" rows with matching Write
func (a *DBUserRepositoryAccess) ByWrite(ctx context.Context, p bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByWrite"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where write = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByWrite: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBUserRepositoryAccess) ByLikeWrite(ctx context.Context, p bool) ([]*savepb.UserRepositoryAccess, error) {
	qn := "DBUserRepositoryAccess_ByLikeWrite"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, userid, read, write from "+a.SQLTablename+" where write ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByWrite: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByWrite: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

func (a *DBUserRepositoryAccess) get_ID(p *savepb.UserRepositoryAccess) uint64 {
	return p.ID
}

func (a *DBUserRepositoryAccess) get_RepoID(p *savepb.UserRepositoryAccess) uint64 {
	return p.RepoID
}

func (a *DBUserRepositoryAccess) get_UserID(p *savepb.UserRepositoryAccess) string {
	return p.UserID
}

func (a *DBUserRepositoryAccess) get_Read(p *savepb.UserRepositoryAccess) bool {
	return p.Read
}

func (a *DBUserRepositoryAccess) get_Write(p *savepb.UserRepositoryAccess) bool {
	return p.Write
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBUserRepositoryAccess) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.UserRepositoryAccess, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBUserRepositoryAccess) Tablename() string {
	return a.SQLTablename
}

func (a *DBUserRepositoryAccess) SelectCols() string {
	return "id,repoid, userid, read, write"
}
func (a *DBUserRepositoryAccess) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repoid, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".read, " + a.SQLTablename + ".write"
}

func (a *DBUserRepositoryAccess) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.UserRepositoryAccess, error) {
	var res []*savepb.UserRepositoryAccess
	for rows.Next() {
		foo := savepb.UserRepositoryAccess{}
		err := rows.Scan(&foo.ID, &foo.RepoID, &foo.UserID, &foo.Read, &foo.Write)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBUserRepositoryAccess) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.UserRepositoryAccess, error) {
	var res []*savepb.UserRepositoryAccess
	for rows.Next() {
		// SCANNER:
		foo := &savepb.UserRepositoryAccess{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepoID
		scanTarget_2 := &foo.UserID
		scanTarget_3 := &foo.Read
		scanTarget_4 := &foo.Write
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4)
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
func (a *DBUserRepositoryAccess) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repoid bigint not null ,userid text not null ,read boolean not null ,write boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repoid bigint not null ,userid text not null ,read boolean not null ,write boolean not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repoid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS read boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS write boolean not null default false;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repoid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS userid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS read boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS write boolean not null  default false;`,
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
func (a *DBUserRepositoryAccess) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

