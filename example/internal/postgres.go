package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	gor "gor/pkg/repository"
)

const (
	userColumns = `id, name, email, created_at`

	// maxPageAlloc caps how many rows one request may preallocate.
	maxPageAlloc = 1000

	sqlCreate = `INSERT INTO users (name, email, created_at)
		VALUES ($1, $2, now())
		RETURNING ` + userColumns

	sqlUpdate = `UPDATE users SET name = $1, email = $2
		WHERE id = $3
		RETURNING ` + userColumns

	sqlFindByID = `SELECT ` + userColumns + ` FROM users WHERE id = $1`

	sqlDelete = `DELETE FROM users WHERE id = $1`

	// count(*) OVER () rides along on the window query: one round trip returns
	// both the page and the total, read from a single consistent snapshot.
	sqlPage = `SELECT ` + userColumns + `, count(*) OVER () AS total
		FROM users ORDER BY id LIMIT $1 OFFSET $2`

	sqlCount = `SELECT count(*) FROM users`
)

// UserRepository stores users in PostgreSQL. Its statements are prepared once
// and reused, so the server parses and plans each query a single time.
type UserRepository struct {
	create   *sql.Stmt
	update   *sql.Stmt
	findByID *sql.Stmt
	delete   *sql.Stmt
	page     *sql.Stmt
	count    *sql.Stmt
}

// NewUserRepository prepares every statement the repository uses. Preparation
// is the one fallible step in construction, hence the error return.
func NewUserRepository(ctx context.Context, db *sql.DB) (*UserRepository, error) {
	r := &UserRepository{}

	for _, s := range []struct {
		dst   **sql.Stmt
		name  string
		query string
	}{
		{&r.create, "create", sqlCreate},
		{&r.update, "update", sqlUpdate},
		{&r.findByID, "findByID", sqlFindByID},
		{&r.delete, "delete", sqlDelete},
		{&r.page, "page", sqlPage},
		{&r.count, "count", sqlCount},
	} {
		stmt, err := db.PrepareContext(ctx, s.query)
		if err != nil {
			// Do not leak the statements already prepared.
			return nil, errors.Join(fmt.Errorf("prepare %s: %w", s.name, err), r.Close())
		}
		*s.dst = stmt
	}

	return r, nil
}

// Close releases the prepared statements. Register it with the engine's
// OnShutdown so they are freed with the rest of the application's resources.
func (r *UserRepository) Close() error {
	var errs []error
	for _, stmt := range []*sql.Stmt{r.create, r.update, r.findByID, r.delete, r.page, r.count} {
		if stmt != nil {
			errs = append(errs, stmt.Close())
		}
	}
	return errors.Join(errs...)
}

func (r *UserRepository) Create(ctx context.Context, u User) (User, error) {
	v, err := scanUser(r.create.QueryRowContext(ctx, u.Name, u.Email))
	if err != nil {
		return ZeroUser, fmt.Errorf("repository.Create: %w", err)
	}
	return v, nil
}

func (r *UserRepository) Update(ctx context.Context, u User) (User, error) {
	v, err := scanUser(r.update.QueryRowContext(ctx, u.Name, u.Email, u.ID))
	if err != nil {
		return ZeroUser, fmt.Errorf("repository.Update: %w", err)
	}
	return v, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (User, error) {
	v, err := scanUser(r.findByID.QueryRowContext(ctx, id))
	if err != nil {
		return ZeroUser, fmt.Errorf("repository.FindByID: %w", err)
	}
	return v, nil
}

func (r *UserRepository) Delete(ctx context.Context, u User) error {
	res, err := r.delete.ExecContext(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("repository.Delete: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.Delete rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("repository.Delete: %w", ErrNotFound)
	}
	return nil
}

func (r *UserRepository) FindAll(ctx context.Context, p gor.PageRequest) (gor.Page[User], error) {
	page, err := r.window(ctx, p)
	if err != nil {
		return page, fmt.Errorf("repository.FindAll: %w", err)
	}
	return page, nil
}

func (r *UserRepository) Search(ctx context.Context, spec gor.SearchSpec) (gor.Page[User], error) {
	panic("not implemented")
}

// window reads one result window plus the total row count in a single query.
func (r *UserRepository) window(ctx context.Context, p gor.PageRequest) (gor.Page[User], error) {
	var page gor.Page[User]

	rows, err := r.page.QueryContext(ctx, p.Limit, p.Offset)
	if err != nil {
		return page, fmt.Errorf("select: %w", err)
	}
	defer rows.Close()

	// Limit is a caller-supplied hint; cap the preallocation so an absurd
	// value cannot make one request reserve unbounded memory.
	page.Data = make([]User, 0, min(max(p.Limit, 0), maxPageAlloc))

	for rows.Next() {
		// Append first, then scan straight into the slice element: no per-row
		// temporary and no copy on append.
		page.Data = append(page.Data, User{})
		u := &page.Data[len(page.Data)-1]
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &page.Total); err != nil {
			return gor.Page[User]{}, fmt.Errorf("row scan: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return gor.Page[User]{}, fmt.Errorf("rows: %w", err)
	}

	// An empty window yields no rows, so no count was read. Offsets past the
	// end are the common case; only then pay for a separate count.
	if len(page.Data) == 0 {
		if err := r.count.QueryRowContext(ctx).Scan(&page.Total); err != nil {
			return gor.Page[User]{}, fmt.Errorf("count: %w", err)
		}
	}

	return page, nil
}

// scanUser reads a single user row, reporting a missing row as ErrNotFound so
// AppErrorHandler maps it to 404.
func scanUser(row *sql.Row) (User, error) {
	var u User
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ZeroUser, ErrNotFound
	}
	if err != nil {
		return ZeroUser, err
	}
	return u, nil
}
