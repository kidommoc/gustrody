package federal

import "github.com/kidommoc/gustrody/internal/models"

type mockingUserAccountDb struct {
	data map[string]struct {
		pub string
		pri string
	}
}

func newMockingUserAccountDb() *mockingUserAccountDb {
	return &mockingUserAccountDb{
		data: make(map[string]struct {
			pub string
			pri string
		}),
	}
}

// never used
func (db *mockingUserAccountDb) SetUser(user *models.User) error {
	return nil
}

func (db *mockingUserAccountDb) QueryUserKeys(username string) (pub string, pri string, err error) {
	return db.data[username].pub, db.data[username].pri, nil
}

// never used
func (db *mockingUserAccountDb) QueryUserPreferences(username string) (pf *models.Preferences, err error) {
	return nil, nil
}

// never used
func (db *mockingUserAccountDb) UpdateUserPreferences(username string, pf *models.Preferences) error {
	return nil
}
