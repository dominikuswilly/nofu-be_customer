package repository

import (
    "errors"
	"strings"
    "log"

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
        `INSERT INTO merchant_master (c_id, c_nm, c_email, c_phone) VALUES ($1, $2, $3, $4)`,
        u.C_ID, u.C_NM, u.C_EMAIL, u.C_PHONE,
    )
    return err
}

func (r *MerchantRepoPG) GetAll() ([]domain.Merchant, error) {
    rows, err := r.db.Query(`SELECT c_id, c_nm, c_email, c_phone FROM merchant_master`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var merchants []domain.Merchant
    for rows.Next() {
        var m domain.Merchant
        var email, phone *string
        if err := rows.Scan(&m.C_ID, &m.C_NM, &email, &phone); err != nil {
            log.Println(err)
            return nil, err
        }
        m.C_EMAIL = email
        m.C_PHONE = phone
        merchants = append(merchants, m)
    }
    return merchants, nil
}

func (r *MerchantRepoPG) GetByID(id string) (*domain.Merchant, error) {
    var m domain.Merchant
    var email, phone *string
    err := r.db.QueryRow(`SELECT c_id, c_nm, c_email, c_phone FROM merchant_master WHERE c_id=$1`, id).
        Scan(&m.C_ID, &m.C_NM, &email, &phone)
    if err != nil {
        return nil, err
    }
    m.C_EMAIL = email
    m.C_PHONE = phone
    return &m, nil
}

func (r *MerchantRepoPG) Update(u *domain.Merchant) error {
    res, err := r.db.Exec(`UPDATE merchant_master SET c_nm=$1, c_email=$2, c_phone=$3 WHERE c_id=$4`,
        u.C_NM, u.C_EMAIL, u.C_PHONE, u.C_ID)
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