package db

/*
 This file was created by mkdb-client.
 The intention is not to modify this file, but you may extend the struct DBSourceRepository
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence sourcerepository_seq;

Main Table:

 CREATE TABLE sourcerepository (id integer primary key default nextval('sourcerepository_seq'),filepath text not null  ,artefactname text not null  ,runpostreceive boolean not null  ,runprereceive boolean not null  ,createdcomplete boolean not null  ,description text not null  ,usercommits bigint not null  ,deleted boolean not null  ,deletedtimestamp integer not null  ,deleteuser text not null  ,lastcommit integer not null  ,lastcommituser text not null  ,tags integer not null  ,forking boolean not null  ,forkedfrom bigint not null  ,buildroutingtagname text not null  ,buildroutingtagvalue text not null  ,readonly boolean not null  ,createuser text not null  ,denymessage text not null  );

Alter statements:
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS filepath text not null default '';
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS artefactname text not null default '';
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS runpostreceive boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS runprereceive boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS createdcomplete boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS description text not null default '';
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS usercommits bigint not null default 0;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS deleted boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS deletedtimestamp integer not null default 0;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS deleteuser text not null default '';
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS lastcommit integer not null default 0;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS lastcommituser text not null default '';
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS tags integer not null default 0;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS forking boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS forkedfrom bigint not null default 0;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS buildroutingtagname text not null default '';
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS buildroutingtagvalue text not null default '';
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS readonly boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS createuser text not null default '';
ALTER TABLE sourcerepository ADD COLUMN IF NOT EXISTS denymessage text not null default '';


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE sourcerepository_archive (id integer unique not null,filepath text not null,artefactname text not null,runpostreceive boolean not null,runprereceive boolean not null,createdcomplete boolean not null,description text not null,usercommits bigint not null,deleted boolean not null,deletedtimestamp integer not null,deleteuser text not null,lastcommit integer not null,lastcommituser text not null,tags integer not null,forking boolean not null,forkedfrom bigint not null,buildroutingtagname text not null,buildroutingtagvalue text not null,readonly boolean not null,createuser text not null,denymessage text not null);
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
	default_def_DBSourceRepository *DBSourceRepository
)

type DBSourceRepository struct {
	DB                   *sql.DB
	SQLTablename         string
	SQLArchivetablename  string
	customColumnHandlers []CustomColumnHandler
	lock                 sync.Mutex
}

func DefaultDBSourceRepository() *DBSourceRepository {
	if default_def_DBSourceRepository != nil {
		return default_def_DBSourceRepository
	}
	psql, err := sql.Open()
	if err != nil {
		fmt.Printf("Failed to open database: %s\n", err)
		os.Exit(10)
	}
	res := NewDBSourceRepository(psql)
	ctx := context.Background()
	err = res.CreateTable(ctx)
	if err != nil {
		fmt.Printf("Failed to create table: %s\n", err)
		os.Exit(10)
	}
	default_def_DBSourceRepository = res
	return res
}
func NewDBSourceRepository(db *sql.DB) *DBSourceRepository {
	foo := DBSourceRepository{DB: db}
	foo.SQLTablename = "sourcerepository"
	foo.SQLArchivetablename = "sourcerepository_archive"
	return &foo
}

func (a *DBSourceRepository) GetCustomColumnHandlers() []CustomColumnHandler {
	return a.customColumnHandlers
}
func (a *DBSourceRepository) AddCustomColumnHandler(w CustomColumnHandler) {
	a.lock.Lock()
	a.customColumnHandlers = append(a.customColumnHandlers, w)
	a.lock.Unlock()
}

func (a *DBSourceRepository) NewQuery() *Query {
	return newQuery(a)
}

// archive. It is NOT transactionally save.
func (a *DBSourceRepository) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBSourceRepository", "insert into "+a.SQLArchivetablename+" (id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom, buildroutingtagname, buildroutingtagvalue, readonly, createuser, denymessage) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) ", p.ID, p.FilePath, p.ArtefactName, p.RunPostReceive, p.RunPreReceive, p.CreatedComplete, p.Description, p.UserCommits, p.Deleted, p.DeletedTimestamp, p.DeleteUser, p.LastCommit, p.LastCommitUser, p.Tags, p.Forking, p.ForkedFrom, p.BuildRoutingTagName, p.BuildRoutingTagValue, p.ReadOnly, p.CreateUser, p.DenyMessage)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// return a map with columnname -> value_from_proto
func (a *DBSourceRepository) buildSaveMap(ctx context.Context, p *savepb.SourceRepository) (map[string]interface{}, error) {
	extra, err := extraFieldsToStore(ctx, a, p)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{})
	res["id"] = a.get_col_from_proto(p, "id")
	res["filepath"] = a.get_col_from_proto(p, "filepath")
	res["artefactname"] = a.get_col_from_proto(p, "artefactname")
	res["runpostreceive"] = a.get_col_from_proto(p, "runpostreceive")
	res["runprereceive"] = a.get_col_from_proto(p, "runprereceive")
	res["createdcomplete"] = a.get_col_from_proto(p, "createdcomplete")
	res["description"] = a.get_col_from_proto(p, "description")
	res["usercommits"] = a.get_col_from_proto(p, "usercommits")
	res["deleted"] = a.get_col_from_proto(p, "deleted")
	res["deletedtimestamp"] = a.get_col_from_proto(p, "deletedtimestamp")
	res["deleteuser"] = a.get_col_from_proto(p, "deleteuser")
	res["lastcommit"] = a.get_col_from_proto(p, "lastcommit")
	res["lastcommituser"] = a.get_col_from_proto(p, "lastcommituser")
	res["tags"] = a.get_col_from_proto(p, "tags")
	res["forking"] = a.get_col_from_proto(p, "forking")
	res["forkedfrom"] = a.get_col_from_proto(p, "forkedfrom")
	res["buildroutingtagname"] = a.get_col_from_proto(p, "buildroutingtagname")
	res["buildroutingtagvalue"] = a.get_col_from_proto(p, "buildroutingtagvalue")
	res["readonly"] = a.get_col_from_proto(p, "readonly")
	res["createuser"] = a.get_col_from_proto(p, "createuser")
	res["denymessage"] = a.get_col_from_proto(p, "denymessage")
	if extra != nil {
		for k, v := range extra {
			res[k] = v
		}
	}
	return res, nil
}

func (a *DBSourceRepository) Save(ctx context.Context, p *savepb.SourceRepository) (uint64, error) {
	qn := "save_DBSourceRepository"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return 0, err
	}
	delete(smap, "id") // save without id
	return a.saveMap(ctx, qn, smap, p)
}

// Save using the ID specified
func (a *DBSourceRepository) SaveWithID(ctx context.Context, p *savepb.SourceRepository) error {
	qn := "insert_DBSourceRepository"
	smap, err := a.buildSaveMap(ctx, p)
	if err != nil {
		return err
	}
	_, err = a.saveMap(ctx, qn, smap, p)
	return err
}

// use a hashmap of columnname->values to store to database (see buildSaveMap())
func (a *DBSourceRepository) saveMap(ctx context.Context, queryname string, smap map[string]interface{}, p *savepb.SourceRepository) (uint64, error) {
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

func (a *DBSourceRepository) Update(ctx context.Context, p *savepb.SourceRepository) error {
	qn := "DBSourceRepository_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set filepath=$1, artefactname=$2, runpostreceive=$3, runprereceive=$4, createdcomplete=$5, description=$6, usercommits=$7, deleted=$8, deletedtimestamp=$9, deleteuser=$10, lastcommit=$11, lastcommituser=$12, tags=$13, forking=$14, forkedfrom=$15, buildroutingtagname=$16, buildroutingtagvalue=$17, readonly=$18, createuser=$19, denymessage=$20 where id = $21", a.get_FilePath(p), a.get_ArtefactName(p), a.get_RunPostReceive(p), a.get_RunPreReceive(p), a.get_CreatedComplete(p), a.get_Description(p), a.get_UserCommits(p), a.get_Deleted(p), a.get_DeletedTimestamp(p), a.get_DeleteUser(p), a.get_LastCommit(p), a.get_LastCommitUser(p), a.get_Tags(p), a.get_Forking(p), a.get_ForkedFrom(p), a.get_BuildRoutingTagName(p), a.get_BuildRoutingTagValue(p), a.get_ReadOnly(p), a.get_CreateUser(p), a.get_DenyMessage(p), p.ID)

	return a.Error(ctx, qn, e)
}

// delete by id field
func (a *DBSourceRepository) DeleteByID(ctx context.Context, p uint64) error {
	qn := "deleteDBSourceRepository_ByID"
	_, e := a.DB.ExecContext(ctx, qn, "delete from "+a.SQLTablename+" where id = $1", p)
	return a.Error(ctx, qn, e)
}

// get it by primary id
func (a *DBSourceRepository) ByID(ctx context.Context, p uint64) (*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, errors.Errorf("No SourceRepository with id %v", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) SourceRepository with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by primary id (nil if no such ID row, but no error either)
func (a *DBSourceRepository) TryByID(ctx context.Context, p uint64) (*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_TryByID"
	l, e := a.fromQuery(ctx, qn, "id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, nil
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, errors.Errorf("Multiple (%d) SourceRepository with id %v", len(l), p))
	}
	return l[0], nil
}

// get it by multiple primary ids
func (a *DBSourceRepository) ByIDs(ctx context.Context, p []uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByIDs"
	l, e := a.fromQuery(ctx, qn, "id in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("TryByID: error scanning (%s)", e))
	}
	return l, nil
}

// get all rows
func (a *DBSourceRepository) All(ctx context.Context) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_all"
	l, e := a.fromQuery(ctx, qn, "true")
	if e != nil {
		return nil, errors.Errorf("All: error scanning (%s)", e)
	}
	return l, nil
}

/**********************************************************************
* GetBy[FIELD] functions
**********************************************************************/

// get all "DBSourceRepository" rows with matching FilePath
func (a *DBSourceRepository) ByFilePath(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByFilePath"
	l, e := a.fromQuery(ctx, qn, "filepath = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFilePath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching FilePath
func (a *DBSourceRepository) ByMultiFilePath(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByFilePath"
	l, e := a.fromQuery(ctx, qn, "filepath in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFilePath: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeFilePath(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeFilePath"
	l, e := a.fromQuery(ctx, qn, "filepath ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByFilePath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching ArtefactName
func (a *DBSourceRepository) ByArtefactName(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByArtefactName"
	l, e := a.fromQuery(ctx, qn, "artefactname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching ArtefactName
func (a *DBSourceRepository) ByMultiArtefactName(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByArtefactName"
	l, e := a.fromQuery(ctx, qn, "artefactname in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeArtefactName(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeArtefactName"
	l, e := a.fromQuery(ctx, qn, "artefactname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByArtefactName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching RunPostReceive
func (a *DBSourceRepository) ByRunPostReceive(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByRunPostReceive"
	l, e := a.fromQuery(ctx, qn, "runpostreceive = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRunPostReceive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching RunPostReceive
func (a *DBSourceRepository) ByMultiRunPostReceive(ctx context.Context, p []bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByRunPostReceive"
	l, e := a.fromQuery(ctx, qn, "runpostreceive in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRunPostReceive: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeRunPostReceive(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeRunPostReceive"
	l, e := a.fromQuery(ctx, qn, "runpostreceive ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRunPostReceive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching RunPreReceive
func (a *DBSourceRepository) ByRunPreReceive(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByRunPreReceive"
	l, e := a.fromQuery(ctx, qn, "runprereceive = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRunPreReceive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching RunPreReceive
func (a *DBSourceRepository) ByMultiRunPreReceive(ctx context.Context, p []bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByRunPreReceive"
	l, e := a.fromQuery(ctx, qn, "runprereceive in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRunPreReceive: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeRunPreReceive(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeRunPreReceive"
	l, e := a.fromQuery(ctx, qn, "runprereceive ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByRunPreReceive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching CreatedComplete
func (a *DBSourceRepository) ByCreatedComplete(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByCreatedComplete"
	l, e := a.fromQuery(ctx, qn, "createdcomplete = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreatedComplete: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching CreatedComplete
func (a *DBSourceRepository) ByMultiCreatedComplete(ctx context.Context, p []bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByCreatedComplete"
	l, e := a.fromQuery(ctx, qn, "createdcomplete in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreatedComplete: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeCreatedComplete(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeCreatedComplete"
	l, e := a.fromQuery(ctx, qn, "createdcomplete ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreatedComplete: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching Description
func (a *DBSourceRepository) ByDescription(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDescription"
	l, e := a.fromQuery(ctx, qn, "description = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching Description
func (a *DBSourceRepository) ByMultiDescription(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDescription"
	l, e := a.fromQuery(ctx, qn, "description in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDescription(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDescription"
	l, e := a.fromQuery(ctx, qn, "description ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching UserCommits
func (a *DBSourceRepository) ByUserCommits(ctx context.Context, p uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByUserCommits"
	l, e := a.fromQuery(ctx, qn, "usercommits = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserCommits: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching UserCommits
func (a *DBSourceRepository) ByMultiUserCommits(ctx context.Context, p []uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByUserCommits"
	l, e := a.fromQuery(ctx, qn, "usercommits in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserCommits: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeUserCommits(ctx context.Context, p uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeUserCommits"
	l, e := a.fromQuery(ctx, qn, "usercommits ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByUserCommits: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching Deleted
func (a *DBSourceRepository) ByDeleted(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeleted"
	l, e := a.fromQuery(ctx, qn, "deleted = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeleted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching Deleted
func (a *DBSourceRepository) ByMultiDeleted(ctx context.Context, p []bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeleted"
	l, e := a.fromQuery(ctx, qn, "deleted in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeleted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDeleted(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDeleted"
	l, e := a.fromQuery(ctx, qn, "deleted ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeleted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching DeletedTimestamp
func (a *DBSourceRepository) ByDeletedTimestamp(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeletedTimestamp"
	l, e := a.fromQuery(ctx, qn, "deletedtimestamp = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeletedTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching DeletedTimestamp
func (a *DBSourceRepository) ByMultiDeletedTimestamp(ctx context.Context, p []uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeletedTimestamp"
	l, e := a.fromQuery(ctx, qn, "deletedtimestamp in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeletedTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDeletedTimestamp(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDeletedTimestamp"
	l, e := a.fromQuery(ctx, qn, "deletedtimestamp ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeletedTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching DeleteUser
func (a *DBSourceRepository) ByDeleteUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeleteUser"
	l, e := a.fromQuery(ctx, qn, "deleteuser = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeleteUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching DeleteUser
func (a *DBSourceRepository) ByMultiDeleteUser(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeleteUser"
	l, e := a.fromQuery(ctx, qn, "deleteuser in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeleteUser: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDeleteUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDeleteUser"
	l, e := a.fromQuery(ctx, qn, "deleteuser ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDeleteUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching LastCommit
func (a *DBSourceRepository) ByLastCommit(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLastCommit"
	l, e := a.fromQuery(ctx, qn, "lastcommit = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastCommit: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching LastCommit
func (a *DBSourceRepository) ByMultiLastCommit(ctx context.Context, p []uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLastCommit"
	l, e := a.fromQuery(ctx, qn, "lastcommit in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastCommit: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeLastCommit(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeLastCommit"
	l, e := a.fromQuery(ctx, qn, "lastcommit ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastCommit: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching LastCommitUser
func (a *DBSourceRepository) ByLastCommitUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLastCommitUser"
	l, e := a.fromQuery(ctx, qn, "lastcommituser = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastCommitUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching LastCommitUser
func (a *DBSourceRepository) ByMultiLastCommitUser(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLastCommitUser"
	l, e := a.fromQuery(ctx, qn, "lastcommituser in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastCommitUser: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeLastCommitUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeLastCommitUser"
	l, e := a.fromQuery(ctx, qn, "lastcommituser ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByLastCommitUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching Tags
func (a *DBSourceRepository) ByTags(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByTags"
	l, e := a.fromQuery(ctx, qn, "tags = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTags: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching Tags
func (a *DBSourceRepository) ByMultiTags(ctx context.Context, p []uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByTags"
	l, e := a.fromQuery(ctx, qn, "tags in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTags: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeTags(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeTags"
	l, e := a.fromQuery(ctx, qn, "tags ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByTags: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching Forking
func (a *DBSourceRepository) ByForking(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByForking"
	l, e := a.fromQuery(ctx, qn, "forking = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByForking: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching Forking
func (a *DBSourceRepository) ByMultiForking(ctx context.Context, p []bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByForking"
	l, e := a.fromQuery(ctx, qn, "forking in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByForking: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeForking(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeForking"
	l, e := a.fromQuery(ctx, qn, "forking ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByForking: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching ForkedFrom
func (a *DBSourceRepository) ByForkedFrom(ctx context.Context, p uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByForkedFrom"
	l, e := a.fromQuery(ctx, qn, "forkedfrom = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByForkedFrom: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching ForkedFrom
func (a *DBSourceRepository) ByMultiForkedFrom(ctx context.Context, p []uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByForkedFrom"
	l, e := a.fromQuery(ctx, qn, "forkedfrom in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByForkedFrom: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeForkedFrom(ctx context.Context, p uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeForkedFrom"
	l, e := a.fromQuery(ctx, qn, "forkedfrom ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByForkedFrom: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching BuildRoutingTagName
func (a *DBSourceRepository) ByBuildRoutingTagName(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByBuildRoutingTagName"
	l, e := a.fromQuery(ctx, qn, "buildroutingtagname = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildRoutingTagName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching BuildRoutingTagName
func (a *DBSourceRepository) ByMultiBuildRoutingTagName(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByBuildRoutingTagName"
	l, e := a.fromQuery(ctx, qn, "buildroutingtagname in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildRoutingTagName: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeBuildRoutingTagName(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeBuildRoutingTagName"
	l, e := a.fromQuery(ctx, qn, "buildroutingtagname ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildRoutingTagName: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching BuildRoutingTagValue
func (a *DBSourceRepository) ByBuildRoutingTagValue(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByBuildRoutingTagValue"
	l, e := a.fromQuery(ctx, qn, "buildroutingtagvalue = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildRoutingTagValue: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching BuildRoutingTagValue
func (a *DBSourceRepository) ByMultiBuildRoutingTagValue(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByBuildRoutingTagValue"
	l, e := a.fromQuery(ctx, qn, "buildroutingtagvalue in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildRoutingTagValue: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeBuildRoutingTagValue(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeBuildRoutingTagValue"
	l, e := a.fromQuery(ctx, qn, "buildroutingtagvalue ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByBuildRoutingTagValue: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching ReadOnly
func (a *DBSourceRepository) ByReadOnly(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByReadOnly"
	l, e := a.fromQuery(ctx, qn, "readonly = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByReadOnly: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching ReadOnly
func (a *DBSourceRepository) ByMultiReadOnly(ctx context.Context, p []bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByReadOnly"
	l, e := a.fromQuery(ctx, qn, "readonly in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByReadOnly: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeReadOnly(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeReadOnly"
	l, e := a.fromQuery(ctx, qn, "readonly ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByReadOnly: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching CreateUser
func (a *DBSourceRepository) ByCreateUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByCreateUser"
	l, e := a.fromQuery(ctx, qn, "createuser = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching CreateUser
func (a *DBSourceRepository) ByMultiCreateUser(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByCreateUser"
	l, e := a.fromQuery(ctx, qn, "createuser in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateUser: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeCreateUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeCreateUser"
	l, e := a.fromQuery(ctx, qn, "createuser ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByCreateUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching DenyMessage
func (a *DBSourceRepository) ByDenyMessage(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDenyMessage"
	l, e := a.fromQuery(ctx, qn, "denymessage = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDenyMessage: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with multiple matching DenyMessage
func (a *DBSourceRepository) ByMultiDenyMessage(ctx context.Context, p []string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDenyMessage"
	l, e := a.fromQuery(ctx, qn, "denymessage in $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDenyMessage: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDenyMessage(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDenyMessage"
	l, e := a.fromQuery(ctx, qn, "denymessage ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, errors.Errorf("ByDenyMessage: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* The field getters
**********************************************************************/

// getter for field "ID" (ID) [uint64]
func (a *DBSourceRepository) get_ID(p *savepb.SourceRepository) uint64 {
	return uint64(p.ID)
}

// getter for field "FilePath" (FilePath) [string]
func (a *DBSourceRepository) get_FilePath(p *savepb.SourceRepository) string {
	return string(p.FilePath)
}

// getter for field "ArtefactName" (ArtefactName) [string]
func (a *DBSourceRepository) get_ArtefactName(p *savepb.SourceRepository) string {
	return string(p.ArtefactName)
}

// getter for field "RunPostReceive" (RunPostReceive) [bool]
func (a *DBSourceRepository) get_RunPostReceive(p *savepb.SourceRepository) bool {
	return bool(p.RunPostReceive)
}

// getter for field "RunPreReceive" (RunPreReceive) [bool]
func (a *DBSourceRepository) get_RunPreReceive(p *savepb.SourceRepository) bool {
	return bool(p.RunPreReceive)
}

// getter for field "CreatedComplete" (CreatedComplete) [bool]
func (a *DBSourceRepository) get_CreatedComplete(p *savepb.SourceRepository) bool {
	return bool(p.CreatedComplete)
}

// getter for field "Description" (Description) [string]
func (a *DBSourceRepository) get_Description(p *savepb.SourceRepository) string {
	return string(p.Description)
}

// getter for field "UserCommits" (UserCommits) [uint64]
func (a *DBSourceRepository) get_UserCommits(p *savepb.SourceRepository) uint64 {
	return uint64(p.UserCommits)
}

// getter for field "Deleted" (Deleted) [bool]
func (a *DBSourceRepository) get_Deleted(p *savepb.SourceRepository) bool {
	return bool(p.Deleted)
}

// getter for field "DeletedTimestamp" (DeletedTimestamp) [uint32]
func (a *DBSourceRepository) get_DeletedTimestamp(p *savepb.SourceRepository) uint32 {
	return uint32(p.DeletedTimestamp)
}

// getter for field "DeleteUser" (DeleteUser) [string]
func (a *DBSourceRepository) get_DeleteUser(p *savepb.SourceRepository) string {
	return string(p.DeleteUser)
}

// getter for field "LastCommit" (LastCommit) [uint32]
func (a *DBSourceRepository) get_LastCommit(p *savepb.SourceRepository) uint32 {
	return uint32(p.LastCommit)
}

// getter for field "LastCommitUser" (LastCommitUser) [string]
func (a *DBSourceRepository) get_LastCommitUser(p *savepb.SourceRepository) string {
	return string(p.LastCommitUser)
}

// getter for field "Tags" (Tags) [uint32]
func (a *DBSourceRepository) get_Tags(p *savepb.SourceRepository) uint32 {
	return uint32(p.Tags)
}

// getter for field "Forking" (Forking) [bool]
func (a *DBSourceRepository) get_Forking(p *savepb.SourceRepository) bool {
	return bool(p.Forking)
}

// getter for field "ForkedFrom" (ForkedFrom) [uint64]
func (a *DBSourceRepository) get_ForkedFrom(p *savepb.SourceRepository) uint64 {
	return uint64(p.ForkedFrom)
}

// getter for field "BuildRoutingTagName" (BuildRoutingTagName) [string]
func (a *DBSourceRepository) get_BuildRoutingTagName(p *savepb.SourceRepository) string {
	return string(p.BuildRoutingTagName)
}

// getter for field "BuildRoutingTagValue" (BuildRoutingTagValue) [string]
func (a *DBSourceRepository) get_BuildRoutingTagValue(p *savepb.SourceRepository) string {
	return string(p.BuildRoutingTagValue)
}

// getter for field "ReadOnly" (ReadOnly) [bool]
func (a *DBSourceRepository) get_ReadOnly(p *savepb.SourceRepository) bool {
	return bool(p.ReadOnly)
}

// getter for field "CreateUser" (CreateUser) [string]
func (a *DBSourceRepository) get_CreateUser(p *savepb.SourceRepository) string {
	return string(p.CreateUser)
}

// getter for field "DenyMessage" (DenyMessage) [string]
func (a *DBSourceRepository) get_DenyMessage(p *savepb.SourceRepository) string {
	return string(p.DenyMessage)
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBSourceRepository) ByDBQuery(ctx context.Context, query *Query) ([]*savepb.SourceRepository, error) {
	extra_fields, err := extraFieldsToQuery(ctx, a)
	if err != nil {
		return nil, err
	}
	i := 0
	for col_name, value := range extra_fields {
		i++
		efname := fmt.Sprintf("EXTRA_FIELD_%d", i)
		query.Add(col_name+" = "+efname, QP{efname: value})
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

func (a *DBSourceRepository) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.SourceRepository, error) {
	return a.fromQuery(ctx, "custom_query_"+a.Tablename(), query_where, args...)
}

// from a query snippet (the part after WHERE)
func (a *DBSourceRepository) fromQuery(ctx context.Context, queryname string, query_where string, args ...interface{}) ([]*savepb.SourceRepository, error) {
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
func (a *DBSourceRepository) get_col_from_proto(p *savepb.SourceRepository, colname string) interface{} {
	if colname == "id" {
		return a.get_ID(p)
	} else if colname == "filepath" {
		return a.get_FilePath(p)
	} else if colname == "artefactname" {
		return a.get_ArtefactName(p)
	} else if colname == "runpostreceive" {
		return a.get_RunPostReceive(p)
	} else if colname == "runprereceive" {
		return a.get_RunPreReceive(p)
	} else if colname == "createdcomplete" {
		return a.get_CreatedComplete(p)
	} else if colname == "description" {
		return a.get_Description(p)
	} else if colname == "usercommits" {
		return a.get_UserCommits(p)
	} else if colname == "deleted" {
		return a.get_Deleted(p)
	} else if colname == "deletedtimestamp" {
		return a.get_DeletedTimestamp(p)
	} else if colname == "deleteuser" {
		return a.get_DeleteUser(p)
	} else if colname == "lastcommit" {
		return a.get_LastCommit(p)
	} else if colname == "lastcommituser" {
		return a.get_LastCommitUser(p)
	} else if colname == "tags" {
		return a.get_Tags(p)
	} else if colname == "forking" {
		return a.get_Forking(p)
	} else if colname == "forkedfrom" {
		return a.get_ForkedFrom(p)
	} else if colname == "buildroutingtagname" {
		return a.get_BuildRoutingTagName(p)
	} else if colname == "buildroutingtagvalue" {
		return a.get_BuildRoutingTagValue(p)
	} else if colname == "readonly" {
		return a.get_ReadOnly(p)
	} else if colname == "createuser" {
		return a.get_CreateUser(p)
	} else if colname == "denymessage" {
		return a.get_DenyMessage(p)
	}
	panic(fmt.Sprintf("in table \"%s\", column \"%s\" cannot be resolved to proto field name", a.Tablename(), colname))
}

func (a *DBSourceRepository) Tablename() string {
	return a.SQLTablename
}

func (a *DBSourceRepository) SelectCols() string {
	return "id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom, buildroutingtagname, buildroutingtagvalue, readonly, createuser, denymessage"
}
func (a *DBSourceRepository) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".filepath, " + a.SQLTablename + ".artefactname, " + a.SQLTablename + ".runpostreceive, " + a.SQLTablename + ".runprereceive, " + a.SQLTablename + ".createdcomplete, " + a.SQLTablename + ".description, " + a.SQLTablename + ".usercommits, " + a.SQLTablename + ".deleted, " + a.SQLTablename + ".deletedtimestamp, " + a.SQLTablename + ".deleteuser, " + a.SQLTablename + ".lastcommit, " + a.SQLTablename + ".lastcommituser, " + a.SQLTablename + ".tags, " + a.SQLTablename + ".forking, " + a.SQLTablename + ".forkedfrom, " + a.SQLTablename + ".buildroutingtagname, " + a.SQLTablename + ".buildroutingtagvalue, " + a.SQLTablename + ".readonly, " + a.SQLTablename + ".createuser, " + a.SQLTablename + ".denymessage"
}

func (a *DBSourceRepository) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.SourceRepository, error) {
	var res []*savepb.SourceRepository
	for rows.Next() {
		// SCANNER:
		foo := &savepb.SourceRepository{}
		// create the non-nullable pointers
		// create variables for scan results
		scanTarget_0 := &foo.ID
		scanTarget_1 := &foo.FilePath
		scanTarget_2 := &foo.ArtefactName
		scanTarget_3 := &foo.RunPostReceive
		scanTarget_4 := &foo.RunPreReceive
		scanTarget_5 := &foo.CreatedComplete
		scanTarget_6 := &foo.Description
		scanTarget_7 := &foo.UserCommits
		scanTarget_8 := &foo.Deleted
		scanTarget_9 := &foo.DeletedTimestamp
		scanTarget_10 := &foo.DeleteUser
		scanTarget_11 := &foo.LastCommit
		scanTarget_12 := &foo.LastCommitUser
		scanTarget_13 := &foo.Tags
		scanTarget_14 := &foo.Forking
		scanTarget_15 := &foo.ForkedFrom
		scanTarget_16 := &foo.BuildRoutingTagName
		scanTarget_17 := &foo.BuildRoutingTagValue
		scanTarget_18 := &foo.ReadOnly
		scanTarget_19 := &foo.CreateUser
		scanTarget_20 := &foo.DenyMessage
		err := rows.Scan(scanTarget_0, scanTarget_1, scanTarget_2, scanTarget_3, scanTarget_4, scanTarget_5, scanTarget_6, scanTarget_7, scanTarget_8, scanTarget_9, scanTarget_10, scanTarget_11, scanTarget_12, scanTarget_13, scanTarget_14, scanTarget_15, scanTarget_16, scanTarget_17, scanTarget_18, scanTarget_19, scanTarget_20)
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
func (a *DBSourceRepository) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),filepath text not null ,artefactname text not null ,runpostreceive boolean not null ,runprereceive boolean not null ,createdcomplete boolean not null ,description text not null ,usercommits bigint not null ,deleted boolean not null ,deletedtimestamp integer not null ,deleteuser text not null ,lastcommit integer not null ,lastcommituser text not null ,tags integer not null ,forking boolean not null ,forkedfrom bigint not null ,buildroutingtagname text not null ,buildroutingtagvalue text not null ,readonly boolean not null ,createuser text not null ,denymessage text not null );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),filepath text not null ,artefactname text not null ,runpostreceive boolean not null ,runprereceive boolean not null ,createdcomplete boolean not null ,description text not null ,usercommits bigint not null ,deleted boolean not null ,deletedtimestamp integer not null ,deleteuser text not null ,lastcommit integer not null ,lastcommituser text not null ,tags integer not null ,forking boolean not null ,forkedfrom bigint not null ,buildroutingtagname text not null ,buildroutingtagvalue text not null ,readonly boolean not null ,createuser text not null ,denymessage text not null );`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS filepath text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS artefactname text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS runpostreceive boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS runprereceive boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createdcomplete boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS description text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS usercommits bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS deleted boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS deletedtimestamp integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS deleteuser text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS lastcommit integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS lastcommituser text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS tags integer not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS forking boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS forkedfrom bigint not null default 0;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS buildroutingtagname text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS buildroutingtagvalue text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS readonly boolean not null default false;`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS createuser text not null default '';`,
		`ALTER TABLE ` + a.SQLTablename + ` ADD COLUMN IF NOT EXISTS denymessage text not null default '';`,

		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS filepath text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS artefactname text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS runpostreceive boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS runprereceive boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createdcomplete boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS description text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS usercommits bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS deleted boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS deletedtimestamp integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS deleteuser text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS lastcommit integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS lastcommituser text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS tags integer not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS forking boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS forkedfrom bigint not null  default 0;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS buildroutingtagname text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS buildroutingtagvalue text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS readonly boolean not null  default false;`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS createuser text not null  default '';`,
		`ALTER TABLE ` + a.SQLTablename + `_archive  ADD COLUMN IF NOT EXISTS denymessage text not null  default '';`,
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
func (a *DBSourceRepository) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return errors.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}

