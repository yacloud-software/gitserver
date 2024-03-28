package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBGroupRepositoryAccess
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence grouprepositoryaccess_seq;

Main Table:

 CREATE TABLE grouprepositoryaccess (id integer primary key default nextval('grouprepositoryaccess_seq'),repoid bigint not null  ,groupid text not null  ,read boolean not null  ,write boolean not null  );

Alter statements:
ALTER TABLE grouprepositoryaccess ADD COLUMN IF NOT EXISTS repoid bigint not null default 0;
ALTER TABLE grouprepositoryaccess ADD COLUMN IF NOT EXISTS groupid text not null default '';
ALTER TABLE grouprepositoryaccess ADD COLUMN IF NOT EXISTS read boolean not null default false;
ALTER TABLE grouprepositoryaccess ADD COLUMN IF NOT EXISTS write boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE grouprepositoryaccess_archive (id integer unique not null,repoid bigint not null,groupid text not null,read boolean not null,write boolean not null);
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
	default_def_DBGroupRepositoryAccess *DBGroupRepositoryAccess
)

type DBGroupRepositoryAccess struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBGroupRepositoryAccess() *DBGroupRepositoryAccess {
	if default_def_DBGroupRepositoryAccess != nil {
		return default_def_DBGroupRepositoryAccess
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBGroupRepositoryAccess(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBGroupRepositoryAccess = res
	return res
}
func NewDBGroupRepositoryAccess(db *sql.DB) *DBGroupRepositoryAccess {
	foo := DBGroupRepositoryAccess{DB: db}
	foo.SQLTablename = "grouprepositoryaccess"
	foo.SQLArchivetablename = "grouprepositoryaccess_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBGroupRepositoryAccess) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBGroupRepositoryAccess", "insert into "+a.SQLArchivetablename+" (id,repoid, groupid, read, write) values ($1,$2, $3, $4, $5) ", p.ID, p.RepoID, p.GroupID, p.Read, p.Write)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBGroupRepositoryAccess) Save(ctx context.Context, p *savepb.GroupRepositoryAccess) (uint64, error) {
	qn := "DBGroupRepositoryAccess_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (repoid, groupid, read, write) values ($1, $2, $3, $4) returning id", a.get_RepoID(p), a.get_GroupID(p), a.get_Read(p), a.get_Write(p))
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
func (a *DBGroupRepositoryAccess) SaveWithID(ctx context.Context, p *savepb.GroupRepositoryAccess) error {
	qn := "insert_DBGroupRepositoryAccess"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,repoid, groupid, read, write) values ($1,$2, $3, $4, $5) ", p.ID, p.RepoID, p.GroupID, p.Read, p.Write)
	return a.Error(ctx, qn, e)
}

func (a *DBGroupRepositoryAccess) Update(ctx context.Context, p *savepb.GroupRepositoryAccess) error {
	qn := "DBGroupRepositoryAccess_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repoid=$1, groupid=$2, read=$3, write=$4 where id = $5", a.get_RepoID(p), a.get_GroupID(p), a.get_Read(p), a.get_Write(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBGroupRepositoryAccess) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBGroupRepositoryAccess_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBGroupRepositoryAccess) ByID(ctx context.Context, p uint64) (*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No GroupRepositoryAccess with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) GroupRepositoryAccess with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBGroupRepositoryAccess) TryByID(ctx context.Context, p uint64) (*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) GroupRepositoryAccess with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBGroupRepositoryAccess) All(ctx context.Context) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" order by id")
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

// get all "DBGroupRepositoryAccess" rows with matching RepoID
func (a *DBGroupRepositoryAccess) ByRepoID(ctx context.Context, p uint64) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByRepoID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where repoid = $1", p)
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
func (a *DBGroupRepositoryAccess) ByLikeRepoID(ctx context.Context, p uint64) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByLikeRepoID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where repoid ilike $1", p)
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

// get all "DBGroupRepositoryAccess" rows with matching GroupID
func (a *DBGroupRepositoryAccess) ByGroupID(ctx context.Context, p string) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByGroupID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where groupid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBGroupRepositoryAccess) ByLikeGroupID(ctx context.Context, p string) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByLikeGroupID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where groupid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByGroupID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBGroupRepositoryAccess" rows with matching Read
func (a *DBGroupRepositoryAccess) ByRead(ctx context.Context, p bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByRead"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where read = $1", p)
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
func (a *DBGroupRepositoryAccess) ByLikeRead(ctx context.Context, p bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByLikeRead"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where read ilike $1", p)
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

// get all "DBGroupRepositoryAccess" rows with matching Write
func (a *DBGroupRepositoryAccess) ByWrite(ctx context.Context, p bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByWrite"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where write = $1", p)
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
func (a *DBGroupRepositoryAccess) ByLikeWrite(ctx context.Context, p bool) ([]*savepb.GroupRepositoryAccess, error) {
	qn := "DBGroupRepositoryAccess_ByLikeWrite"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repoid, groupid, read, write from "+a.SQLTablename+" where write ilike $1", p)
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

func (a *DBGroupRepositoryAccess) get_ID(p *savepb.GroupRepositoryAccess) uint64 {
	return p.ID
}

func (a *DBGroupRepositoryAccess) get_RepoID(p *savepb.GroupRepositoryAccess) uint64 {
	return p.RepoID
}

func (a *DBGroupRepositoryAccess) get_GroupID(p *savepb.GroupRepositoryAccess) string {
	return p.GroupID
}

func (a *DBGroupRepositoryAccess) get_Read(p *savepb.GroupRepositoryAccess) bool {
	return p.Read
}

func (a *DBGroupRepositoryAccess) get_Write(p *savepb.GroupRepositoryAccess) bool {
	return p.Write
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBGroupRepositoryAccess) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.GroupRepositoryAccess, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBGroupRepositoryAccess) Tablename() string {
	return a.SQLTablename
}

func (a *DBGroupRepositoryAccess) SelectCols() string {
	return "id,repoid, groupid, read, write"
}
func (a *DBGroupRepositoryAccess) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repoid, " + a.SQLTablename + ".groupid, " + a.SQLTablename + ".read, " + a.SQLTablename + ".write"
}

func (a *DBGroupRepositoryAccess) FromRowsOld(ctx context.Context, rows *gosql.Rows) ([]*savepb.GroupRepositoryAccess, error) {
	var res []*savepb.GroupRepositoryAccess
	for rows.Next() {
		foo := savepb.GroupRepositoryAccess{}
		err := rows.Scan(&foo.ID, &foo.RepoID, &foo.GroupID, &foo.Read, &foo.Write)
		if err != nil {
			return nil, a.Error(ctx, "fromrow-scan", err)
		}
		res = append(res, &foo)
	}
	return res, nil
}
func (a *DBGroupRepositoryAccess) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.GroupRepositoryAccess, error) {
	var res []*savepb.GroupRepositoryAccess
	for rows.Next() {
		// SCANNER:
		foo := &savepb.GroupRepositoryAccess{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepoID
		scanTarget_2 := &foo.GroupID
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
func (a *DBGroupRepositoryAccess) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repoid bigint not null ,groupid text not null ,read boolean not null ,write boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repoid bigint not null ,groupid text not null ,read boolean not null ,write boolean not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS repoid bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS groupid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS read boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS write boolean not null default false;`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS repoid bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS groupid text not null  default '';`,
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
func (a *DBGroupRepositoryAccess) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

