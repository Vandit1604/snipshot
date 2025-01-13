package mysql

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/vandit1604/snipshot/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

// We'll use the Insert method to add a new record to the users table.
func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users(name, email, hashed_password,created) VALUES(?,?,?,UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		// here we're type asserting the err and if it is correct then we're using the struct values to check what kind of mysql error was thrown
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				return models.ErrDuplicateEmail
			}
		}
	}

	return nil
}

// Authenticate verifies a user exists with the provided email address and password.
// If the user exists the relevant user ID is returned.
func (u *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPw []byte
	stmt := `SELECT id, hashed_password FROM users WHERE email = ?`
	row := u.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.ErrInvalidCredenetials
		} else {
			return 0, err
		}
	}

	// Check whether the hashed password and plain-text password provided match.
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPw, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, models.ErrInvalidCredenetials
		} else {
			return 0, err
		}
	}

	// Otherwise, the password is correct, so return the userID.
	return id, nil
}

// We'll use the Get method to fetch details for a specific user based
// on their user ID.
func (m *UserModel) Get(id int) (*models.User, error) {
	user := &models.User{}

	stmt := `SELECT id,name,email,created FROM users WHERE id=?`

	err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.Created)
	if err == models.ErrRecordNotFound {
		return nil, models.ErrRecordNotFound
	} else if err != nil {
		return nil, err
	}

	return user, nil
}
