package db

/*
 This file was created by mkdb-client.
 The intention is not to modify thils file, but you may extend the struct DBSourceRepository
 in a seperate file (so that you can regenerate this one from time to time)
*/

/*
 PRIMARY KEY: ID
*/

/*
 postgres:
 create sequence sourcerepository_seq;

Main Table:

 CREATE TABLE sourcerepository (id integer primary key default nextval('sourcerepository_seq'),filepath text not null  ,artefactname text not null  ,runpostreceive boolean not null  ,runprereceive boolean not null  ,createdcomplete boolean not null  ,description text not null  ,usercommits bigint not null  ,deleted boolean not null  ,deletedtimestamp integer not null  ,deleteuser text not null  ,lastcommit integer not null  ,lastcommituser text not null  ,tags integer not null  ,forking boolean not null  ,forkedfrom bigint not null  );

Alter statements:
ALTER TABLE sourcerepository ADD COLUMN filepath text not null default '';
ALTER TABLE sourcerepository ADD COLUMN artefactname text not null default '';
ALTER TABLE sourcerepository ADD COLUMN runpostreceive boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN runprereceive boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN createdcomplete boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN description text not null default '';
ALTER TABLE sourcerepository ADD COLUMN usercommits bigint not null default 0;
ALTER TABLE sourcerepository ADD COLUMN deleted boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN deletedtimestamp integer not null default 0;
ALTER TABLE sourcerepository ADD COLUMN deleteuser text not null default '';
ALTER TABLE sourcerepository ADD COLUMN lastcommit integer not null default 0;
ALTER TABLE sourcerepository ADD COLUMN lastcommituser text not null default '';
ALTER TABLE sourcerepository ADD COLUMN tags integer not null default 0;
ALTER TABLE sourcerepository ADD COLUMN forking boolean not null default false;
ALTER TABLE sourcerepository ADD COLUMN forkedfrom bigint not null default 0;


Archive Table: (structs can be moved from main to archive using Archive() function)

 CREATE TABLE sourcerepository_archive (id integer unique not null,filepath text not null,artefactname text not null,runpostreceive boolean not null,runprereceive boolean not null,createdcomplete boolean not null,description text not null,usercommits bigint not null,deleted boolean not null,deletedtimestamp integer not null,deleteuser text not null,lastcommit integer not null,lastcommituser text not null,tags integer not null,forking boolean not null,forkedfrom bigint not null);
*/

import (
	"context"
	gosql "database/sql"
	"fmt"
	savepb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/go-easyops/sql"
)

type DBSourceRepository struct {
	DB                  *sql.DB
	SQLTablename        string
	SQLArchivetablename string
}

func NewDBSourceRepository(db *sql.DB) *DBSourceRepository {
	foo := DBSourceRepository{DB: db}
	foo.SQLTablename = "sourcerepository"
	foo.SQLArchivetablename = "sourcerepository_archive"
	return &foo
}

// archive. It is NOT transactionally save.
func (a *DBSourceRepository) Archive(ctx context.Context, id uint64) error {

	// load it
	p, err := a.ByID(ctx, id)
	if err != nil {
		return err
	}

	// now save it to archive:
	_, e := a.DB.ExecContext(ctx, "archive_DBSourceRepository", "insert into "+a.SQLArchivetablename+"+ (id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) ", p.ID, p.FilePath, p.ArtefactName, p.RunPostReceive, p.RunPreReceive, p.CreatedComplete, p.Description, p.UserCommits, p.Deleted, p.DeletedTimestamp, p.DeleteUser, p.LastCommit, p.LastCommitUser, p.Tags, p.Forking, p.ForkedFrom)
	if e != nil {
		return e
	}

	// now delete it.
	a.DeleteByID(ctx, id)
	return nil
}

// Save (and use database default ID generation)
func (a *DBSourceRepository) Save(ctx context.Context, p *savepb.SourceRepository) (uint64, error) {
	qn := "DBSourceRepository_Save"
	rows, e := a.DB.QueryContext(ctx, qn, "insert into "+a.SQLTablename+" (filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) returning id", p.FilePath, p.ArtefactName, p.RunPostReceive, p.RunPreReceive, p.CreatedComplete, p.Description, p.UserCommits, p.Deleted, p.DeletedTimestamp, p.DeleteUser, p.LastCommit, p.LastCommitUser, p.Tags, p.Forking, p.ForkedFrom)
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
func (a *DBSourceRepository) SaveWithID(ctx context.Context, p *savepb.SourceRepository) error {
	qn := "insert_DBSourceRepository"
	_, e := a.DB.ExecContext(ctx, qn, "insert into "+a.SQLTablename+" (id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom) values ($1,$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) ", p.ID, p.FilePath, p.ArtefactName, p.RunPostReceive, p.RunPreReceive, p.CreatedComplete, p.Description, p.UserCommits, p.Deleted, p.DeletedTimestamp, p.DeleteUser, p.LastCommit, p.LastCommitUser, p.Tags, p.Forking, p.ForkedFrom)
	return a.Error(ctx, qn, e)
}

func (a *DBSourceRepository) Update(ctx context.Context, p *savepb.SourceRepository) error {
	qn := "DBSourceRepository_Update"
	_, e := a.DB.ExecContext(ctx, qn, "update "+a.SQLTablename+" set filepath=$1, artefactname=$2, runpostreceive=$3, runprereceive=$4, createdcomplete=$5, description=$6, usercommits=$7, deleted=$8, deletedtimestamp=$9, deleteuser=$10, lastcommit=$11, lastcommituser=$12, tags=$13, forking=$14, forkedfrom=$15 where id = $16", p.FilePath, p.ArtefactName, p.RunPostReceive, p.RunPreReceive, p.CreatedComplete, p.Description, p.UserCommits, p.Deleted, p.DeletedTimestamp, p.DeleteUser, p.LastCommit, p.LastCommitUser, p.Tags, p.Forking, p.ForkedFrom, p.ID)

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
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where id = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByID: error scanning (%s)", e))
	}
	if len(l) == 0 {
		return nil, a.Error(ctx, qn, fmt.Errorf("No SourceRepository with id %d", p))
	}
	if len(l) != 1 {
		return nil, a.Error(ctx, qn, fmt.Errorf("Multiple (%d) SourceRepository with id %d", len(l), p))
	}
	return l[0], nil
}

// get all rows
func (a *DBSourceRepository) All(ctx context.Context) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_all"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" order by id")
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

// get all "DBSourceRepository" rows with matching FilePath
func (a *DBSourceRepository) ByFilePath(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByFilePath"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where filepath = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFilePath: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFilePath: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeFilePath(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeFilePath"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where filepath ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFilePath: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByFilePath: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching ArtefactName
func (a *DBSourceRepository) ByArtefactName(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByArtefactName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where artefactname = $1", p)
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
func (a *DBSourceRepository) ByLikeArtefactName(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeArtefactName"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where artefactname ilike $1", p)
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

// get all "DBSourceRepository" rows with matching RunPostReceive
func (a *DBSourceRepository) ByRunPostReceive(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByRunPostReceive"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where runpostreceive = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRunPostReceive: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRunPostReceive: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeRunPostReceive(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeRunPostReceive"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where runpostreceive ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRunPostReceive: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRunPostReceive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching RunPreReceive
func (a *DBSourceRepository) ByRunPreReceive(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByRunPreReceive"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where runprereceive = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRunPreReceive: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRunPreReceive: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeRunPreReceive(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeRunPreReceive"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where runprereceive ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRunPreReceive: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByRunPreReceive: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching CreatedComplete
func (a *DBSourceRepository) ByCreatedComplete(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByCreatedComplete"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where createdcomplete = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreatedComplete: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreatedComplete: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeCreatedComplete(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeCreatedComplete"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where createdcomplete ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreatedComplete: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByCreatedComplete: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching Description
func (a *DBSourceRepository) ByDescription(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDescription"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where description = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDescription(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDescription"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where description ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDescription: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching UserCommits
func (a *DBSourceRepository) ByUserCommits(ctx context.Context, p uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByUserCommits"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where usercommits = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserCommits: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserCommits: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeUserCommits(ctx context.Context, p uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeUserCommits"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where usercommits ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserCommits: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByUserCommits: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching Deleted
func (a *DBSourceRepository) ByDeleted(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeleted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where deleted = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeleted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeleted: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDeleted(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDeleted"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where deleted ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeleted: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeleted: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching DeletedTimestamp
func (a *DBSourceRepository) ByDeletedTimestamp(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeletedTimestamp"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where deletedtimestamp = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeletedTimestamp: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeletedTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDeletedTimestamp(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDeletedTimestamp"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where deletedtimestamp ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeletedTimestamp: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeletedTimestamp: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching DeleteUser
func (a *DBSourceRepository) ByDeleteUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByDeleteUser"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where deleteuser = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeleteUser: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeleteUser: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeDeleteUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeDeleteUser"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where deleteuser ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeleteUser: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByDeleteUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching LastCommit
func (a *DBSourceRepository) ByLastCommit(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLastCommit"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where lastcommit = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastCommit: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastCommit: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeLastCommit(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeLastCommit"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where lastcommit ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastCommit: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastCommit: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching LastCommitUser
func (a *DBSourceRepository) ByLastCommitUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLastCommitUser"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where lastcommituser = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastCommitUser: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastCommitUser: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeLastCommitUser(ctx context.Context, p string) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeLastCommitUser"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where lastcommituser ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastCommitUser: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByLastCommitUser: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching Tags
func (a *DBSourceRepository) ByTags(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByTags"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where tags = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTags: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTags: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeTags(ctx context.Context, p uint32) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeTags"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where tags ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTags: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByTags: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching Forking
func (a *DBSourceRepository) ByForking(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByForking"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where forking = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByForking: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByForking: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeForking(ctx context.Context, p bool) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeForking"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where forking ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByForking: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByForking: error scanning (%s)", e))
	}
	return l, nil
}

// get all "DBSourceRepository" rows with matching ForkedFrom
func (a *DBSourceRepository) ByForkedFrom(ctx context.Context, p uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByForkedFrom"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where forkedfrom = $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByForkedFrom: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByForkedFrom: error scanning (%s)", e))
	}
	return l, nil
}

// the 'like' lookup
func (a *DBSourceRepository) ByLikeForkedFrom(ctx context.Context, p uint64) ([]*savepb.SourceRepository, error) {
	qn := "DBSourceRepository_ByLikeForkedFrom"
	rows, e := a.DB.QueryContext(ctx, qn, "select id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom from "+a.SQLTablename+" where forkedfrom ilike $1", p)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByForkedFrom: error querying (%s)", e))
	}
	defer rows.Close()
	l, e := a.FromRows(ctx, rows)
	if e != nil {
		return nil, a.Error(ctx, qn, fmt.Errorf("ByForkedFrom: error scanning (%s)", e))
	}
	return l, nil
}

/**********************************************************************
* Helper to convert from an SQL Query
**********************************************************************/

// from a query snippet (the part after WHERE)
func (a *DBSourceRepository) FromQuery(ctx context.Context, query_where string, args ...interface{}) ([]*savepb.SourceRepository, error) {
	rows, err := a.DB.QueryContext(ctx, "custom_query_"+a.Tablename(), "select "+a.SelectCols()+" from "+a.Tablename()+" where "+query_where, args...)
	if err != nil {
		return nil, err
	}
	return a.FromRows(ctx, rows)
}

/**********************************************************************
* Helper to convert from an SQL Row to struct
**********************************************************************/
func (a *DBSourceRepository) Tablename() string {
	return a.SQLTablename
}

func (a *DBSourceRepository) SelectCols() string {
	return "id,filepath, artefactname, runpostreceive, runprereceive, createdcomplete, description, usercommits, deleted, deletedtimestamp, deleteuser, lastcommit, lastcommituser, tags, forking, forkedfrom"
}
func (a *DBSourceRepository) SelectColsQualified() string {
	return "" + a.SQLTablename + ".id," + a.SQLTablename + ".filepath, " + a.SQLTablename + ".artefactname, " + a.SQLTablename + ".runpostreceive, " + a.SQLTablename + ".runprereceive, " + a.SQLTablename + ".createdcomplete, " + a.SQLTablename + ".description, " + a.SQLTablename + ".usercommits, " + a.SQLTablename + ".deleted, " + a.SQLTablename + ".deletedtimestamp, " + a.SQLTablename + ".deleteuser, " + a.SQLTablename + ".lastcommit, " + a.SQLTablename + ".lastcommituser, " + a.SQLTablename + ".tags, " + a.SQLTablename + ".forking, " + a.SQLTablename + ".forkedfrom"
}

func (a *DBSourceRepository) FromRows(ctx context.Context, rows *gosql.Rows) ([]*savepb.SourceRepository, error) {
	var res []*savepb.SourceRepository
	for rows.Next() {
		foo := savepb.SourceRepository{}
		err := rows.Scan(&foo.ID, &foo.FilePath, &foo.ArtefactName, &foo.RunPostReceive, &foo.RunPreReceive, &foo.CreatedComplete, &foo.Description, &foo.UserCommits, &foo.Deleted, &foo.DeletedTimestamp, &foo.DeleteUser, &foo.LastCommit, &foo.LastCommitUser, &foo.Tags, &foo.Forking, &foo.ForkedFrom)
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
func (a *DBSourceRepository) CreateTable(ctx context.Context) error {
	csql := []string{
		`create sequence if not exists ` + a.SQLTablename + `_seq;`,
		`CREATE TABLE if not exists ` + a.SQLTablename + ` (id integer primary key default nextval('` + a.SQLTablename + `_seq'),filepath text not null  ,artefactname text not null  ,runpostreceive boolean not null  ,runprereceive boolean not null  ,createdcomplete boolean not null  ,description text not null  ,usercommits bigint not null  ,deleted boolean not null  ,deletedtimestamp integer not null  ,deleteuser text not null  ,lastcommit integer not null  ,lastcommituser text not null  ,tags integer not null  ,forking boolean not null  ,forkedfrom bigint not null  );`,
		`CREATE TABLE if not exists ` + a.SQLTablename + `_archive (id integer primary key default nextval('` + a.SQLTablename + `_seq'),filepath text not null  ,artefactname text not null  ,runpostreceive boolean not null  ,runprereceive boolean not null  ,createdcomplete boolean not null  ,description text not null  ,usercommits bigint not null  ,deleted boolean not null  ,deletedtimestamp integer not null  ,deleteuser text not null  ,lastcommit integer not null  ,lastcommituser text not null  ,tags integer not null  ,forking boolean not null  ,forkedfrom bigint not null  );`,
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
func (a *DBSourceRepository) Error(ctx context.Context, q string, e error) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("[table="+a.SQLTablename+", query=%s] Error: %s", q, e)
}
