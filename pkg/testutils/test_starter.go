package testutils

func NewTestAppServer() *AppServer {
	db, closer := NewTestDB()
	return &AppServer{
		DB:       db,
		dbCloser: closer,
	}
}
