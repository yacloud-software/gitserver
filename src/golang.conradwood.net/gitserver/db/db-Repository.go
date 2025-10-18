package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBRepository
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
ALTER TABLE repository ADD COLUMN IF NOT EXISTS reponame text not null default '';
ALTER TABLE repository ADD COLUMN IF NOT EXISTS ownerid text not null default '';
ALTER TABLE repository ADD COLUMN IF NOT EXISTS artefactname text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE repository_archive (id integer unique not null,reponame text not null,ownerid text not null,artefactname text not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/go-easyops/sql"
	"os"
	"sync"
)

var (
	default_def_DBRepository *DBRepository
)

type DBRepository struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func init() {
	RegisterDBHandlerFactory(func() Handler {
		return DefaultDBRepository()
	})
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

func (a *DBRepository) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBRepository) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBRepository) NewQuery() *Query {
	return newQuery(a)
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

// return a map with columnname -> value_from_proto
func (a *DBRepository) buildSaveMap(ctx context.Context, p *savepb.Repository) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["reponame"] = a.get_col_from_proto(p, "reponame")
	res["ownerid"] = a.get_col_from_proto(p, "ownerid")
	res["artefactname"] = a.get_col_from_proto(p, "artefactname")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBRepository) Save(ctx context.Context, p *savepb.Repository) (uint64, error) {
	qn := "save_DBRepository"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBRepository) SaveWithID(ctx context.Context, p *savepb.Repository) error {
	qn := "insert_DBRepository"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBRepository) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.Repository) (uint64, error) {
	// Save (and use database default ID generation)

	var rows *gosql.Rows
	var e error

	q_cols := ""
	q_valnames := ""
	q_vals := make([]interface{}, 0)
	deli := ""
	i := 0
	// build the 2 parts of the query (column names and value names) as well as the values themselves
	for colname, val := range smap {
		q_cols = q_cols + deli + colname
		i++
		q_valnames = q_valnames + deli + fmt.Sprintf("$%d", i)
		q_vals = append(q_vals, val)
		deli = ","
	}
	rows, e = a.DB.QueryContext(ctx, queryname, "insert into "+a.SQLTablename+" ("+q_cols+") values ("+q_valnames+") returning id", q_vals...)
	if e != nil {
		return 0, a.Error(ctx, queryname, e)
	}
	defer rows.Close()
	if !rows.Next() {
		return 0, a.Error(ctx, queryname, errors.Errorf("No rows after insert"))
	}
	var id uint64
	e = rows.Scan(&id)
	if e != nil {
		return 0, a.Error(ctx, queryname, errors.Errorf("failed to scan id after insert: %s", e))
	}
	p.ID = id
	return id, nil
}

// if ID==0 save, otherwise update
func (a *DBRepository) SaveOrUpdate(ctx context.Context, p *savepb.Repository) error {
	if p.ID == 0 {
		_, err := a.Save(ctx, p)
		return err
	}
	return a.Update(ctx, p)
}
func (a *DBRepository) Update(ctx context.Context, p *savepb.Repository) error {
	qn := "DBRepository_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set reponame=$1, ownerid=$2, artefactname=$3 where id = $4", a.get_RepoName(p), a.get_OwnerID(p), a.get_ArtefactName(p), p.ID)

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
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No Repository with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Repository with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBRepository) TryByID(ctx context.Context, p uint64) (*savepb.Repository, error) {
	qn := "DBRepository_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) Repository with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBRepository) ByIDs(ctx context.Context, p []uint64) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBRepository) All(ctx context.Context) ([]*savepb.Repository, error) {
	qn := "DBRepository_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBRepository" rows with matching RepoName
func (a *DBRepository) ByRepoName(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByRepoName"
	l, e := a.fromQuery(ctx, qn, "reponame = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepository" rows with multiple matching RepoName
func (a *DBRepository) ByMultiRepoName(ctx context.Context, p []string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByRepoName"
	l, e := a.fromQuery(ctx, qn, "reponame in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepository) ByLikeRepoName(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByLikeRepoName"
	l, e := a.fromQuery(ctx, qn, "reponame ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRepoName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepository" rows with matching OwnerID
func (a *DBRepository) ByOwnerID(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByOwnerID"
	l, e := a.fromQuery(ctx, qn, "ownerid = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByOwnerID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepository" rows with multiple matching OwnerID
func (a *DBRepository) ByMultiOwnerID(ctx context.Context, p []string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByOwnerID"
	l, e := a.fromQuery(ctx, qn, "ownerid in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByOwnerID: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepository) ByLikeOwnerID(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByLikeOwnerID"
	l, e := a.fromQuery(ctx, qn, "ownerid ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByOwnerID: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepository" rows with matching ArtefactName
func (a *DBRepository) ByArtefactName(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByArtefactName"
	l, e := a.fromQuery(ctx, qn, "artefactname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBRepository" rows with multiple matching ArtefactName
func (a *DBRepository) ByMultiArtefactName(ctx context.Context, p []string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByArtefactName"
	l, e := a.fromQuery(ctx, qn, "artefactname in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBRepository) ByLikeArtefactName(ctx context.Context, p string) ([]*savepb.Repository, error) {
	qn := "DBRepository_ByLikeArtefactName"
	l, e := a.fromQuery(ctx, qn, "artefactname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactName: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBRepository) get_ID(p *savepb.Repository) uint64 {
	return uint64(p.ID)
}

// getter for field "RepoName" (RepoName) [string]
func (a *DBRepository) get_RepoName(p *savepb.Repository) string {
	return string(p.RepoName)
}

// getter for field "OwnerID" (OwnerID) [string]
func (a *DBRepository) get_OwnerID(p *savepb.Repository) string {
	return string(p.OwnerID)
}

// getter for field "ArtefactName" (ArtefactName) [string]
func (a *DBRepository) get_ArtefactName(p *savepb.Repository) string {
	return string(p.ArtefactName)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBRepository) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.Repository, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		/*
		   efname:=fmt.Sprintf("EXTRA_FIELD_%d",i)
		   query.Add(col_name+" = "+efname,QP{efname:value})
		*/
		query.AddEqual(col_name, value)
	}

	gw, paras := query.ToPostgres()
	queryname := "custom_dbquery"
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where "+gw, paras...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (a *DBRepository) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.Repository, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBRepository) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.Repository, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	eq := ""
	if extra_fields != nil && len(extra_fields) > 0 {
		eq = " AND ("
		// build the extraquery "eq"
		i := len(args)
		deli := ""
		for col_name, value := range extra_fields {
			i++
			eq = eq + deli + col_name + fmt.Sprintf(" = $%d", i)
			deli = " AND "
			args = append(args, value)
		}
		eq = eq + ")"
	}
	rows, err := a.DB.QueryContext(ctx, queryname, "select "+a.SelectCols()+" from "+a.Tablename()+" where ( "+query_where+") "+eq, args...)
	if err != nil {
		return nil, err
	}
	res, err := a.FromRows(ctx, rows)
	rows.Close()
	if err != nil {
		return nil, err
	}
	return res, nil
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBRepository) get_col_from_proto(p *savepb.Repository, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "reponame" {
		return a.get_RepoName(p)
	} else if colname == "ownerid" {
		return a.get_OwnerID(p)
	} else if colname == "artefactname" {
		return a.get_ArtefactName(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

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
		// SCANNER:
		foo := &savepb.Repository{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.RepoName
		scanTarget_2 := &foo.OwnerID
		scanTarget_3 := &foo.ArtefactName
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3)
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
func (a *DBRepository) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),reponame text not null ,ownerid text not null ,artefactname text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),reponame text not null ,ownerid text not null ,artefactname text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS reponame text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS ownerid text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS artefactname text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS reponame text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS ownerid text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS artefactname text not null  default '';`,
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
func (a *DBRepository) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

