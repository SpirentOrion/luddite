package main

import (
	"encoding/xml"
	"net/http"
	"sync"

	log "github.com/SpirentOrion/logrus"
	"github.com/SpirentOrion/luddite"
)

type User struct {
	XMLName  xml.Name `json:"-" xml:"user"`
	Name     string   `json:"name" xml:"name" schema:"name"`
	Password string   `json:"password,omitempty" xml:"password,omitempty" schema:"password"`
}

func (u *User) SafeExport() *User {
	return &User{
		Name: u.Name,
	}
}

type userResource struct {
	sync.RWMutex
	users map[string]*User
}

func newUserResource() *userResource {
	return &userResource{users: make(map[string]*User)}
}

func (r *userResource) New() interface{} {
	return &User{}
}

func (r *userResource) Id(value interface{}) string {
	u := value.(*User)
	return u.Name
}

func (r *userResource) List(req *http.Request) (int, interface{}) {
	r.RLock()
	defer r.RUnlock()

	us := make([]*User, len(r.users), len(r.users))
	idx := 0
	for _, u := range r.users {
		us[idx] = u.SafeExport()
		idx++
	}
	return http.StatusOK, us
}

func (r *userResource) Get(req *http.Request, id string) (int, interface{}) {
	r.RLock()
	defer r.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return http.StatusNotFound, nil
	}
	return http.StatusOK, u.SafeExport()
}

func (r *userResource) Create(req *http.Request, value interface{}) (int, interface{}) {
	u := value.(*User)
	r.Lock()
	defer r.Unlock()

	_, exists := r.users[u.Name]
	if exists {
		return http.StatusBadRequest, luddite.NewError(errorDefs, EcodeUserExists, u.Name)
	}
	r.users[u.Name] = u

	logger := luddite.ContextLogger(req.Context())
	logger.WithFields(log.Fields{
		"event": "UserCreated",
		"name":  u.Name,
	}).Info()

	return http.StatusCreated, u.SafeExport()
}

func (r *userResource) Update(req *http.Request, id string, value interface{}) (int, interface{}) {
	u := value.(*User)
	r.Lock()
	defer r.Unlock()

	_, exists := r.users[id]
	if !exists {
		return http.StatusNotFound, nil
	}
	r.users[u.Name] = u

	logger := luddite.ContextLogger(req.Context())
	logger.WithFields(log.Fields{
		"event": "UserUpdated",
		"name":  u.Name,
	}).Info()

	return http.StatusOK, u.SafeExport()
}

func (r *userResource) Delete(req *http.Request, id string) (int, interface{}) {
	r.Lock()
	defer r.Unlock()

	delete(r.users, id)

	logger := luddite.ContextLogger(req.Context())
	logger.WithFields(log.Fields{
		"event": "UserDeleted",
		"name":  id,
	}).Info()

	return http.StatusNoContent, nil
}

func (r *userResource) Action(req *http.Request, id string, action string) (int, interface{}) {
	switch action {
	case "reset_password":
		return r.resetPassword(req, id)
	default:
		return http.StatusNotFound, nil
	}
}

func (r *userResource) resetPassword(req *http.Request, id string) (int, interface{}) {
	r.Lock()
	defer r.Unlock()

	u, exists := r.users[id]
	if !exists {
		return http.StatusNotFound, nil
	}
	u.Password = "secret"

	logger := luddite.ContextLogger(req.Context())
	logger.WithFields(log.Fields{
		"event": "UserPasswordReset",
		"name":  id,
	}).Info()

	return http.StatusNoContent, nil
}
