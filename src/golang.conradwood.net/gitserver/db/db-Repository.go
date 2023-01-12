package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBRepository
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence repository_seq;

Main Table:

 CREATE TABLE repository (id integer primary key default nextval('repository_seq'),reponame text not null  ,ownerid text not null  ,artefactname text not null  );

Alter statements:
ALTER TABLE repository ADD COLUMN reponame text not null default '';
ALTER TABLE repository ADD COLUMN ownerid text not null default '';
ALTER TABLE repository ADD COLUMN artefactname text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE repository_archive (id integer unique not null,reponame text not null,ownerid text not null,artefactname text not null);
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
	default_def_DBRepository *DBRepository
)

type DBRepository struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func DefaultDBRepository() *DBRepository {
	if default_def_DBRepository != nil {
		return default_def_DBRepository
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBRepository(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBRepository = res
	return res
}
func NewDBRepository(db *sql.DB) *DBRepository {
	foo := DBRepository{DB: db}
	foo.SQLTablename = "repository"
	foo.SQLArchivetablename = "repository_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBRepository) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBRepository", "insert into "+a.SQLArchivetablename+" (id,reponame, ownerid, artefactname) values ($1,$2, $3, $4) ", p.ID, p.RepoName, p.OwnerID, p.ArtefactName)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBRepository) Save(ctx context.Context, p *savepb.Repository) (uint64, error) {
	qn := "DBRepository_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (reponame, ownerid, artefactname) values ($1, $2, $3) returning id", p.RepoName, p.OwnerID, p.ArtefactName)
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
func (a *DBRepository) SaveWithID(ctx context.Context, p *savepb.Repository) error {
	qn := "insert_DBRepository"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,reponame, ownerid, artefactname) values ($1,$2, $3, $4) ", p.ID, p.RepoName, p.OwnerID, p.ArtefactName)
	return a.Error(ctx, qn, e)
}

func (a *DBRepository) Update(ctx context.Context, p *savepb.Repository) error {
	qn := "DBRepository_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set reponame=$1, ownerid=$2, artefactname=$3 where id = $4", p.RepoName, p.OwnerID, p.ArtefactName, p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBRepository) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBRepository_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBRepository) ByID(ctx context.Context, p uint64) (*savepb.Repository, error) {
	qn := "DBRepository_ByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No Repository with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Repository with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBRepository) TryByID(ctx context.Context, p uint64) (*savepb.Repository, error) {
	qn := "DBRepository_TryByID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" where id = $1", p)
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
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) Repository with id %v", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBRepository) All(ctx context.Context) ([]*savepb.Repository, error) {
	qn := "DBRepository_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" order by id")
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

// get all "DBRepository" rows with matching RepoName
func (a *DBRepository) ByRepoName(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByRepoName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" where reponame = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepository) ByLikeRepoName(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByLikeRepoName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" where reponame ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepository" rows with matching OwnerID
func (a *DBRepository) ByOwnerID(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByOwnerID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" where ownerid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOwnerID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOwnerID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepository) ByLikeOwnerID(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByLikeOwnerID"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" where ownerid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOwnerID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByOwnerID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepository" rows with matching ArtefactName
func (a *DBRepository) ByArtefactName(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByArtefactName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" where artefactname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArtefactName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArtefactName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepository) ByLikeArtefactName(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByLikeArtefactName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,reponame, ownerid, artefactname from "+a.SQLTablename+" where artefactname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArtefactName: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByArtefactName: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBRepository) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Repository, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBRepository) Tablename() string {
	return a.SQLTablename
}

func (a *DBRepository) SelectCols() string {
	return "id,reponame, ownerid, artefactname"
}
func (a *DBRepository) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".reponame, " + a.SQLTablename + ".ownerid, " + a.SQLTablename + ".artefactname"
}

func (a *DBRepository) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.Repository, error) {
	var res []*savepb.Repository
	for rows.Next() {
		foo := savepb.Repository{Permission: &savepb.Permission{}}
		err := rows.Scan(&foo.ID, &foo.RepoName, &foo.OwnerID, &foo.ArtefactName)
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
func (a *DBRepository) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),reponame text not null  ,ownerid text not null  ,artefactname text not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),reponame text not null  ,ownerid text not null  ,artefactname text not null  );`,
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
func (a *DBRepository) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
