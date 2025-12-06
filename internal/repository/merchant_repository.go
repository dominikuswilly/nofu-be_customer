package repository

import (
    "errors"
	"strings"

    "database/sql" 
    _ "github.com/jackc/pgx/v5/stdlib"

    "github.com/google/uuid"
    "be_customer/internal/domain"
)

type MerchantRepoPG struct {
    db *sql.DB
}

func NewMerchantRepoPG(db *sql.DB) *MerchantRepoPG {
    return &MerchantRepoPG{db: db}
}

func (r *MerchantRepoPG) Create(u *domain.Merchant) error {
    v7, err := uuid.NewV7()
    if err != nil {
        return err
    }
    u.C_ID = strings.ReplaceAll(v7.String(), "-", "")

    _, err = r.db.Exec(
        `INSERT INTO merchant_master (c_id, c_nm) VALUES ($1, $2)`,
        u.C_ID, u.C_NM,
    )
    return err
}

func (r *MerchantRepoPG) GetAll() ([]domain.Merchant, error) {
    rows, err := r.db.Query(`SELECT c_id, c_nm FROM merchant_master`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var merchants []domain.Merchant
    for rows.Next() {
        var m domain.Merchant
        if err := rows.Scan(&m.C_ID, &m.C_NM); err != nil {
            return nil, err
        }
        merchants = append(merchants, m)
    }
    return merchants, nil
}

func (r *MerchantRepoPG) GetByID(id string) (*domain.Merchant, error) {
    var m domain.Merchant
    err := r.db.QueryRow(`SELECT c_id, c_nm FROM merchant_master WHERE c_id=$1`, id).
        Scan(&m.C_ID, &m.C_NM)
    if err != nil {
        return nil, err
    }
    return &m, nil
}

func (r *MerchantRepoPG) Update(u *domain.Merchant) error {
    res, err := r.db.Exec(`UPDATE merchant_master SET c_nm=$1 WHERE c_id=$2`,
        u.C_NM, u.C_ID)
    if err != nil {
        return err
    }
    count, _ := res.RowsAffected()
    if count == 0 {
        return errors.New("merchant not found")
    }
    return nil
}

func (r *MerchantRepoPG) Delete(id string) error {
    res, err := r.db.Exec(`DELETE FROM merchant_master WHERE c_id=$1`, id)
    if err != nil {
        return err
    }
    count, _ := res.RowsAffected()
    if count == 0 {
        return errors.New("merchant not found")
    }
    return nil
}