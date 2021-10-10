package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBSourceRepositoryURL
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence sourcerepositoryurl_seq;

Main Table:

 CREATE TABLE sourcerepositoryurl (id integer primary key default nextval('sourcerepositoryurl_seq'),v2repositoryid bigint not null  ,host text not null  ,path text not null  );

Alter statements:
ALTER TABLE sourcerepositoryurl ADD COLUMN v2repositoryid bigint not null default 0;
ALTER TABLE sourcerepositoryurl ADD COLUMN host text not null default '';
ALTER TABLE sourcerepositoryurl ADD COLUMN path text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE sourcerepositoryurl_archive (id integer unique not null,v2repositoryid bigint not null,host text not null,path text not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/sql"
)

type DBSourceRepositoryURL struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func NewDBSourceRepositoryURL(db *sql.DB) *DBSourceRepositoryURL {
	foo := DBSourceRepositoryURL{DB: db}
	foo.SQLTablename = "sourcerepositoryurl"
	foo.SQLArchivetablename = "sourcerepositoryurl_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBSourceRepositoryURL) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBSourceRepositoryURL", "insert into "+a.SQLArchivetablename+"+ (id,v2repositoryid, host, path) values ($1,$2, $3, $4) ", p.ID, p.V2RepositoryID, p.Host, p.Path)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBSourceRepositoryURL) Save(ctx context.Context, p *savepb.SourceRepositoryURL) (uint64, error) {
	qn := "DBSourceRepositoryURL_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (v2repositoryid, host, path) values ($1, $2, $3) returning id", p.V2RepositoryID, p.Host, p.Path)
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
func (a *DBSourceRepositoryURL) SaveWithID(ctx context.Context, p *savepb.SourceRepositoryURL) error {
	qn := "insert_DBSourceRepositoryURL"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,v2repositoryid, host, path) values ($1,$2, $3, $4) ", p.ID, p.V2RepositoryID, p.Host, p.Path)
	return a.Error(ctx, qn, e)
}

func (a *DBSourceRepositoryURL) Update(ctx context.Context, p *savepb.SourceRepositoryURL) error {
	qn := "DBSourceRepositoryURL_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set v2repositoryid=$1, host=$2, path=$3 where id = $4", p.V2RepositoryID, p.Host, p.Path, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBSourceRepositoryURL) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBSourceRepositoryURL_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBSourceRepositoryURL) ByID(ctx context.Context, p uint64) (*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,v2repositoryid, host, path from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No SourceRepositoryURL with id %d", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) SourceRepositoryURL with id %d", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBSourceRepositoryURL) All(ctx context.Context) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,v2repositoryid, host, path from "+a.SQLTablename+" order by id")
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

// get all "DBSourceRepositoryURL" rows with matching V2RepositoryID
func (a *DBSourceRepositoryURL) ByV2RepositoryID(ctx context.Context, p uint64) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByV2RepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,v2repositoryid, host, path from "+a.SQLTablename+" where v2repositoryid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByV2RepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByV2RepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepositoryURL) ByLikeV2RepositoryID(ctx context.Context, p uint64) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByLikeV2RepositoryID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,v2repositoryid, host, path from "+a.SQLTablename+" where v2repositoryid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByV2RepositoryID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByV2RepositoryID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepositoryURL" rows with matching Host
func (a *DBSourceRepositoryURL) ByHost(ctx context.Context, p string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,v2repositoryid, host, path from "+a.SQLTablename+" where host = $1", p)
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
func (a *DBSourceRepositoryURL) ByLikeHost(ctx context.Context, p string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByLikeHost"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,v2repositoryid, host, path from "+a.SQLTablename+" where host ilike $1", p)
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

// get all "DBSourceRepositoryURL" rows with matching Path
func (a *DBSourceRepositoryURL) ByPath(ctx context.Context, p string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByPath"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,v2repositoryid, host, path from "+a.SQLTablename+" where path = $1", p)
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
func (a *DBSourceRepositoryURL) ByLikePath(ctx context.Context, p string) ([]*savepb.SourceRepositoryURL, error) {
	qn := "DBSourceRepositoryURL_ByLikePath"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,v2repositoryid, host, path from "+a.SQLTablename+" where path ilike $1", p)
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

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBSourceRepositoryURL) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.SourceRepositoryURL, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBSourceRepositoryURL) Tablename() string {
	return a.SQLTablename
}

func (a *DBSourceRepositoryURL) SelectCols() string {
	return "id,v2repositoryid, host, path"
}
func (a *DBSourceRepositoryURL) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".v2repositoryid, " + a.SQLTablename + ".host, " + a.SQLTablename + ".path"
}

func (a *DBSourceRepositoryURL) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.SourceRepositoryURL, error) {
	var res []*savepb.SourceRepositoryURL
	for rows.Next() {
		foo := savepb.SourceRepositoryURL{}
		err := rows.Scan(&foo.ID, &foo.V2RepositoryID, &foo.Host, &foo.Path)
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
func (a *DBSourceRepositoryURL) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),v2repositoryid bigint not null  ,host text not null  ,path text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),v2repositoryid bigint not null  ,host text not null  ,path text not null  );`,
	}
	for i, c := range csql {
		_, e := a.DB.ExecContext(ctx, fmt.Sprintf("create_"+a.SQLTablename+"_%d", i), c)
		if e != nil {
			return e
		}
	}
	return nil
}

/**********************************************************************
* Helper to meaningful errors
**********************************************************************/
func (a *DBSourceRepositoryURL) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
