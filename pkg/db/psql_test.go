package db

import (
	"reflect"
	"testing"

	"github.com/fattymango/px-take-home/config"
	"github.com/stretchr/testify/require"
)

func TestNewMysqDB(t *testing.T) {

	testCases := []struct {
		Name          string
		Config        *config.Config
		CheckResponse func(t *testing.T, d *DB, err error)
	}{
		{
			Name: "OK",
			Config: &config.Config{
				DB: config.DB{
					Host:     "localhost",
					Port:     "5432",
					User:     "crm",
					Password: "crm.123",
					Name:     "postgres",
					SSLMode:  "disable",
				},
			},
			CheckResponse: func(t *testing.T, d *DB, err error) {
				require.NoError(t, err)
				require.NotNil(t, d)
			},
		},
		{
			Name: "InvalidURL",
			Config: &config.Config{
				DB: config.DB{
					Host:     "localhost",
					Port:     "5432",
					User:     "crm",
					Password: "crm.123",
					Name:     "postgres",
					SSLMode:  "disable",
				},
			},
			CheckResponse: func(t *testing.T, d *DB, err error) {
				require.Error(t, err)
				require.Nil(t, d)
			},
		},
		{
			Name: "InvalidCredentials",
			Config: &config.Config{
				DB: config.DB{
					Host:     "localhost",
					Port:     "5432",
					User:     "invalid",
					Password: "invalid",
					Name:     "invalid",
					SSLMode:  "disable",
				},
			},
			CheckResponse: func(t *testing.T, d *DB, err error) {
				require.Error(t, err)
				require.Nil(t, d)
			},
		},
		{
			Name: "InvalidDBName",
			Config: &config.Config{
				DB: config.DB{
					Host:     "localhost",
					Port:     "5432",
					User:     "crm",
					Password: "crm.123",
					Name:     "invalid",
					SSLMode:  "disable",
				},
			},
			CheckResponse: func(t *testing.T, d *DB, err error) {
				require.Error(t, err)
				require.Nil(t, d)
			},
		},
	}

	for i := range testCases {
		t.Run(testCases[i].Name, func(t *testing.T) {
			db, err := NewSQLiteDB(testCases[i].Config)
			testCases[i].CheckResponse(t, db, err)
		})
	}
}

func TestNewTestPsqlDB(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "OK",
			args: args{
				cfg: &config.Config{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewSQLiteDB(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTestPsqlDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			db.AutoMigrate(&User{})
			db.Create(&User{ID: 1, Name: "John"})
			u := User{}
			err = db.Where("id = ?", 1).First(&u).Error
			if err != nil {
				t.Errorf("NewTestPsqlDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(u, User{ID: 1, Name: "John"}) {
				t.Errorf("NewTestPsqlDB() = %v, want %v", u, User{ID: 1, Name: "John"})
			}

		})
	}
}
