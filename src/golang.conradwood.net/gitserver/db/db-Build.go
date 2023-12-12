package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBBuild
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence build_seq;

Main Table:

 CREATE TABLE build (id integer primary key default nextval('build_seq'),repositoryid bigint not null  ,commithash text not null  ,branch text not null  ,logmessage text not null  ,userid text not null  ,r_timestamp integer not null  ,success boolean not null  );

Alter statements:
ALTER TABLE build ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;
ALTER TABLE build ADD COLUMN IF NOT EXISTS commithash text not null default '';
ALTER TABLE build ADD COLUMN IF NOT EXISTS branch text not null default '';
ALTER TABLE build ADD COLUMN IF NOT EXISTS logmessage text not null default '';
ALTER TABLE build ADD COLUMN IF NOT EXISTS userid text not null default '';
ALTER TABLE build ADD COLUMN IF NOT EXISTS r_timestamp integer not null default 0;
ALTER TABLE build ADD COLUMN IF NOT EXISTS success boolean not null default false;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE build_archive (id integer unique not null,repositoryid bigint not null,commithash text not null,branch text not null,logmessage text not null,userid text not null,r_timestamp integer not null,success boolean not null);
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
	default_def_DBBuild *DBBuild
)

type DBBuild struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBBuild() *DBBuild {
	if default_def_DBBuild != nil {
		return default_def_DBBuild
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBBuild(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBBuild = res
	return res
}
func NewDBBuild(db *sql.DB) *DBBuild {
	foo := DBBuild{DB: db}
	foo.SQLTablename = "build"
	foo.SQLArchivetablename = "build_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBBuild) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBBuild", "insert into "+a.SQLArchivetablename+" (id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success) values ($1,$2, $3, $4, $5, $6, $7, $8) ", p.ID, p.RepositoryID, p.CommitHash, p.Branch, p.LogMessage, p.UserID, p.Timestamp, p.Success)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBBuild) Save(ctx context.Context, p *savepb.Build) (uint64, error) {
	qn := "DBBuild_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (repositoryid, commithash, branch, logmessage, userid, r_timestamp, success) values ($1, $2, $3, $4, $5, $6, $7) returning id", p.RepositoryID, p.CommitHash, p.Branch, p.LogMessage, p.UserID, p.Timestamp, p.Success)
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
func (a *DBBuild) SaveWithID(ctx context.Context, p *savepb.Build) error {
	qn := "insert_DBBuild"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success) values ($1,$2, $3, $4, $5, $6, $7, $8) ", p.ID, p.RepositoryID, p.CommitHash, p.Branch, p.LogMessage, p.UserID, p.Timestamp, p.Success)
	return a.Error(ctx, qn, e)
}

func (a *DBBuild) Update(ctx context.Context, p *savepb.Build) error {
	qn := "DBBuild_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set repositoryid=$1, commithash=$2, branch=$3, logmessage=$4, userid=$5, r_timestamp=$6, success=$7 where id = $8", p.RepositoryID, p.CommitHash, p.Branch, p.LogMessage, p.UserID, p.Timestamp, p.Success, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBBuild) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBBuild_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBBuild) ByID(ctx context.Context, p uint64) (*savepb.Build, error) {
	qn := "DBBuild_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Build with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Build with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBBuild) TryByID(ctx context.Context, p uint64) (*savepb.Build, error) {
	qn := "DBBuild_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Build with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBBuild) All(ctx context.Context) ([]*savepb.Build, error) {
	qn := "DBBuild_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" order by id")
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

// get all "DBBuild" rows with matching RepositoryID
func (a *DBBuild) ByRepositoryID(ctx context.Context, p uint64) ([]*savepb.Build, error) {
	qn := "DBBuild_ByRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeRepositoryID(ctx context.Context, p uint64) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeRepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching CommitHash
func (a *DBBuild) ByCommitHash(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByCommitHash"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where commithash = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCommitHash: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCommitHash: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeCommitHash(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeCommitHash"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where commithash ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCommitHash: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCommitHash: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching Branch
func (a *DBBuild) ByBranch(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByBranch"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where branch = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBranch: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBranch: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeBranch(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeBranch"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where branch ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBranch: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByBranch: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching LogMessage
func (a *DBBuild) ByLogMessage(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLogMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where logmessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLogMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeLogMessage(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeLogMessage"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where logmessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLogMessage: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLogMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching UserID
func (a *DBBuild) ByUserID(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where userid = $1", p)
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
func (a *DBBuild) ByLikeUserID(ctx context.Context, p string) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeUserID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where userid ilike $1", p)
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

// get all "DBBuild" rows with matching Timestamp
func (a *DBBuild) ByTimestamp(ctx context.Context, p uint32) ([]*savepb.Build, error) {
	qn := "DBBuild_ByTimestamp"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where r_timestamp = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTimestamp: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeTimestamp(ctx context.Context, p uint32) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeTimestamp"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where r_timestamp ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTimestamp: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBBuild" rows with matching Success
func (a *DBBuild) BySuccess(ctx context.Context, p bool) ([]*savepb.Build, error) {
	qn := "DBBuild_BySuccess"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where success = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBBuild) ByLikeSuccess(ctx context.Context, p bool) ([]*savepb.Build, error) {
	qn := "DBBuild_ByLikeSuccess"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success from "+a.SQLTablename+" where success ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("BySuccess: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBBuild) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Build, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBBuild) Tablename() string {
	return a.SQLTablename
}

func (a *DBBuild) SelectCols() string {
	return "id,repositoryid, commithash, branch, logmessage, userid, r_timestamp, success"
}
func (a *DBBuild) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".repositoryid, " + a.SQLTablename + ".commithash, " + a.SQLTablename + ".branch, " + a.SQLTablename + ".logmessage, " + a.SQLTablename + ".userid, " + a.SQLTablename + ".r_timestamp, " + a.SQLTablename + ".success"
}

func (a *DBBuild) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Build, error) {
	var res []*savepb.Build
	for rows.Next() {
		foo := savepb.Build{}
		err := rows.Scan(&foo.ID, &foo.RepositoryID, &foo.CommitHash, &foo.Branch, &foo.LogMessage, &foo.UserID, &foo.Timestamp, &foo.Success)
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
func (a *DBBuild) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,commithash text not null ,branch text not null ,logmessage text not null ,userid text not null ,r_timestamp integer not null ,success boolean not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),repositoryid bigint not null ,commithash text not null ,branch text not null ,logmessage text not null ,userid text not null ,r_timestamp integer not null ,success boolean not null );`,
		`ALTER TABLE build ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE build ADD COLUMN IF NOT EXISTS commithash text not null default '';`,
		`ALTER TABLE build ADD COLUMN IF NOT EXISTS branch text not null default '';`,
		`ALTER TABLE build ADD COLUMN IF NOT EXISTS logmessage text not null default '';`,
		`ALTER TABLE build ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE build ADD COLUMN IF NOT EXISTS r_timestamp integer not null default 0;`,
		`ALTER TABLE build ADD COLUMN IF NOT EXISTS success boolean not null default false;`,

		`ALTER TABLE build_archive ADD COLUMN IF NOT EXISTS repositoryid bigint not null default 0;`,
		`ALTER TABLE build_archive ADD COLUMN IF NOT EXISTS commithash text not null default '';`,
		`ALTER TABLE build_archive ADD COLUMN IF NOT EXISTS branch text not null default '';`,
		`ALTER TABLE build_archive ADD COLUMN IF NOT EXISTS logmessage text not null default '';`,
		`ALTER TABLE build_archive ADD COLUMN IF NOT EXISTS userid text not null default '';`,
		`ALTER TABLE build_archive ADD COLUMN IF NOT EXISTS r_timestamp integer not null default 0;`,
		`ALTER TABLE build_archive ADD COLUMN IF NOT EXISTS success boolean not null default false;`,
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
func (a *DBBuild) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}



