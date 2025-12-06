package repository

import (
    "errors"
	"strings"

    "database/sql" 
    _ "github.com/jackc/pgx/v5/stdlib"

    "github.com/google/uuid"
    "be_customer/internal/domain"
)

type AdminRepoPG struct {
    db *sql.DB
}

func NewAdminRepoPG(db *sql.DB) *AdminRepoPG {
    return &AdminRepoPG{db: db}
}

func (r *AdminRepoPG) Create(u *domain.Admin) error {
    v7, err := uuid.NewV7()
    if err != nil {
        return err
    }
    u.C_ID = strings.ReplaceAll(v7.String(), "-", "")

    _, err = r.db.Exec(
        `INSERT INTO admin_master (c_id, c_nm, c_email, c_phone) VALUES ($1, $2, $3, $4)`,
        u.C_ID, u.C_NM, u.C_EMAIL, u.C_PHONE,
    )
    return err
}

func (r *AdminRepoPG) GetAll() ([]domain.Admin, error) {
    rows, err := r.db.Query(`SELECT c_id, c_nm, c_email, c_phone FROM admin_master`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var admins []domain.Admin
    for rows.Next() {
        var a domain.Admin
        if err := rows.Scan(&a.C_ID, &a.C_NM, &a.C_EMAIL, &a.C_PHONE); err != nil {
            return nil, err
        }
        admins = append(admins, a)
    }
    return admins, nil
}

func (r *AdminRepoPG) GetByID(id string) (*domain.Admin, error) {
    var a domain.Admin
    err := r.db.QueryRow(`SELECT c_id, c_nm, c_email, c_phone FROM admin_master WHERE c_id=$1`, id).
        Scan(&a.C_ID, &a.C_NM, &a.C_EMAIL, &a.C_PHONE)
    if err != nil {
        return nil, err
    }
    return &a, nil
}

func (r *AdminRepoPG) Update(u *domain.Admin) error {
    res, err := r.db.Exec(`UPDATE admin_master SET c_nm=$1 WHERE c_id=$2`,
        u.C_NM, u.C_ID)
    if err != nil {
        return err
    }
    count, _ := res.RowsAffected()
    if count == 0 {
        return errors.New("admin not found")
    }
    return nil
}

func (r *AdminRepoPG) Delete(id string) error {
    res, err := r.db.Exec(`DELETE FROM admin_master WHERE c_id=$1`, id)
    if err != nil {
        return err
    }
    count, _ := res.RowsAffected()
    if count == 0 {
        return errors.New("admin not found")
    }
    return nil
}