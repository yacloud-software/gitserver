package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBPingState
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence pingstate_seq;

Main Table:

 CREATE TABLE pingstate (id integer primary key default nextval('pingstate_seq'),associationtoken text not null  ,created integer not null  ,responsetoken text not null  );

Alter statements:
ALTER TABLE pingstate ADD COLUMN associationtoken text not null default '';
ALTER TABLE pingstate ADD COLUMN created integer not null default 0;
ALTER TABLE pingstate ADD COLUMN responsetoken text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE pingstate_archive (id integer unique not null,associationtoken text not null,created integer not null,responsetoken text not null);
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
	default_def_DBPingState *DBPingState
)

type DBPingState struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBPingState() *DBPingState {
	if default_def_DBPingState != nil {
		return default_def_DBPingState
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBPingState(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBPingState = res
	return res
}
func NewDBPingState(db *sql.DB) *DBPingState {
	foo := DBPingState{DB: db}
	foo.SQLTablename = "pingstate"
	foo.SQLArchivetablename = "pingstate_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBPingState) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBPingState", "insert into "+a.SQLArchivetablename+" (id,associationtoken, created, responsetoken) values ($1,$2, $3, $4) ", p.ID, p.AssociationToken, p.Created, p.ResponseToken)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBPingState) Save(ctx context.Context, p *savepb.PingState) (uint64, error) {
	qn := "DBPingState_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (associationtoken, created, responsetoken) values ($1, $2, $3) returning id", p.AssociationToken, p.Created, p.ResponseToken)
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
func (a *DBPingState) SaveWithID(ctx context.Context, p *savepb.PingState) error {
	qn := "insert_DBPingState"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,associationtoken, created, responsetoken) values ($1,$2, $3, $4) ", p.ID, p.AssociationToken, p.Created, p.ResponseToken)
	return a.Error(ctx, qn, e)
}

func (a *DBPingState) Update(ctx context.Context, p *savepb.PingState) error {
	qn := "DBPingState_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set associationtoken=$1, created=$2, responsetoken=$3 where id = $4", p.AssociationToken, p.Created, p.ResponseToken, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBPingState) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBPingState_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBPingState) ByID(ctx context.Context, p uint64) (*savepb.PingState, error) {
	qn := "DBPingState_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,associationtoken, created, responsetoken from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No PingState with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) PingState with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBPingState) All(ctx context.Context) ([]*savepb.PingState, error) {
	qn := "DBPingState_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,associationtoken, created, responsetoken from "+a.SQLTablename+" order by id")
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

// get all "DBPingState" rows with matching AssociationToken
func (a *DBPingState) ByAssociationToken(ctx context.Context, p string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByAssociationToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,associationtoken, created, responsetoken from "+a.SQLTablename+" where associationtoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAssociationToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingState) ByLikeAssociationToken(ctx context.Context, p string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByLikeAssociationToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,associationtoken, created, responsetoken from "+a.SQLTablename+" where associationtoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAssociationToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByAssociationToken: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingState" rows with matching Created
func (a *DBPingState) ByCreated(ctx context.Context, p uint32) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,associationtoken, created, responsetoken from "+a.SQLTablename+" where created = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingState) ByLikeCreated(ctx context.Context, p uint32) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByLikeCreated"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,associationtoken, created, responsetoken from "+a.SQLTablename+" where created ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreated: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreated: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBPingState" rows with matching ResponseToken
func (a *DBPingState) ByResponseToken(ctx context.Context, p string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByResponseToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,associationtoken, created, responsetoken from "+a.SQLTablename+" where responsetoken = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByResponseToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByResponseToken: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBPingState) ByLikeResponseToken(ctx context.Context, p string) ([]*savepb.PingState, error) {
	qn := "DBPingState_ByLikeResponseToken"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,associationtoken, created, responsetoken from "+a.SQLTablename+" where responsetoken ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByResponseToken: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByResponseToken: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBPingState) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.PingState, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBPingState) Tablename() string {
	return a.SQLTablename
}

func (a *DBPingState) SelectCols() string {
	return "id,associationtoken, created, responsetoken"
}
func (a *DBPingState) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".associationtoken, " + a.SQLTablename + ".created, " + a.SQLTablename + ".responsetoken"
}

func (a *DBPingState) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.PingState, error) {
	var res []*savepb.PingState
	for rows.Next() {
		foo := savepb.PingState{}
		err := rows.Scan(&foo.ID, &foo.AssociationToken, &foo.Created, &foo.ResponseToken)
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
func (a *DBPingState) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),associationtoken text not null  ,created integer not null  ,responsetoken text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),associationtoken text not null  ,created integer not null  ,responsetoken text not null  );`,
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
func (a *DBPingState) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
