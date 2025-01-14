package macro

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/reef-pi/reef-pi/controller/utils"
)

func (t *Subsystem) LoadAPI(r *mux.Router) {
	r.HandleFunc("/api/macros", t.list).Methods("GET")
	r.HandleFunc("/api/macros/scheduled", t.scheduledlist).Methods("GET")
	r.HandleFunc("/api/macros", t.create).Methods("PUT")
	r.HandleFunc("/api/macros/{id}", t.get).Methods("GET")
	r.HandleFunc("/api/macros/{id}", t.update).Methods("POST")
	r.HandleFunc("/api/macros/{id}", t.delete).Methods("DELETE")
	r.HandleFunc("/api/macros/{id}/run", t.run).Methods("POST")
	r.HandleFunc("/api/macros/{id}/revert", t.revert).Methods("POST")
}

func (t *Subsystem) get(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) (interface{}, error) {
		return t.Get(id)
	}
	utils.JSONGetResponse(fn, w, r)
}

func (c Subsystem) list(w http.ResponseWriter, r *http.Request) {
	fn := func() (interface{}, error) {
		return c.List()
	}
	utils.JSONListResponse(fn, w, r)
}
func (c *Subsystem) scheduledlist(w http.ResponseWriter, r *http.Request) {
	fn := func() (interface{}, error) {
		return c.scheduledmacros, nil
	}
	utils.JSONListResponse(fn, w, r)
}

func (c *Subsystem) create(w http.ResponseWriter, r *http.Request) {
	var m Macro
	fn := func() error {
		return c.Create(m)
	}
	utils.JSONCreateResponse(&m, fn, w, r)
}

func (c *Subsystem) delete(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) error {
		return c.Delete(id)
	}
	utils.JSONDeleteResponse(fn, w, r)
}

func (c *Subsystem) update(w http.ResponseWriter, r *http.Request) {
	var m Macro
	fn := func(id string) error {
		return c.Update(id, m)
	}
	utils.JSONUpdateResponse(&m, fn, w, r)
}

func (c *Subsystem) run(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) error {
		m, err := c.Get(id)
		if err != nil {
			return err
		}
		go c.Run(m, false)
		return nil
	}
	utils.JSONDeleteResponse(fn, w, r)
}

func (c *Subsystem) revert(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) error {
		m, err := c.Get(id)
		if err != nil {
			return err
		}
		if !m.Reversible {
			return errors.New("macro is not reversible")
		}
		go c.Run(m, true)
		return nil
	}
	utils.JSONDeleteResponse(fn, w, r)
}
