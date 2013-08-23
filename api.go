package gawker

import (
    "encoding/json"
    "fmt"
    "github.com/fathomdb/processes"
    "github.com/gorilla/mux"
    "log"
    "net"
    "net/http"
    "os"
    "strconv"
    "strings"
)

const DEFAULTHTTPHOST string = "127.0.0.1"

//const DEFAULTHTTPPORT int = 4243

func httpError(w http.ResponseWriter, err error) {
    if strings.HasPrefix(err.Error(), "No such") {
        http.Error(w, err.Error(), http.StatusNotFound)
    } else if strings.HasPrefix(err.Error(), "Bad parameter") {
        http.Error(w, err.Error(), http.StatusBadRequest)
    } else if strings.HasPrefix(err.Error(), "Conflict") {
        http.Error(w, err.Error(), http.StatusConflict)
    } else if strings.HasPrefix(err.Error(), "Impossible") {
        http.Error(w, err.Error(), http.StatusNotAcceptable)
    } else if strings.HasPrefix(err.Error(), "Wrong login/password") {
        http.Error(w, err.Error(), http.StatusUnauthorized)
    } else if strings.Contains(err.Error(), "hasn't been activated") {
        http.Error(w, err.Error(), http.StatusForbidden)
    } else {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

//If we don't do this, POST method without Content-type (even with empty body) will fail
func parseForm(r *http.Request) error {
    if err := r.ParseForm(); err != nil && !strings.HasPrefix(err.Error(), "mime:") {
        return err
    }
    return nil
}

func writeJSON(w http.ResponseWriter, b []byte) {
    w.Header().Set("Content-Type", "application/json")
    w.Write(b)
}

func getBoolParam(value string) (bool, error) {
    if value == "" {
        return false, nil
    }
    ret, err := strconv.ParseBool(value)
    if err != nil {
        return false, fmt.Errorf("Bad parameter")
    }
    return ret, nil
}

func getProcesses(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    if err := parseForm(r); err != nil {
        return err
    }

    outs := srv.GetProcesses()
    b, err := json.Marshal(outs)
    if err != nil {
        return err
    }

    writeJSON(w, b)
    return nil
}

func postProcess(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    config := &processes.WatchedProcessConfig{}

    if err := json.NewDecoder(r.Body).Decode(config); err != nil {
        return err
    }

    name := vars["name"]

    err := srv.Runtime.Processes.WriteProcess(name, config)
    if err != nil {
        return err
    }
    //	out.ID = id

    //	b, err := json.Marshal(out)
    //	if err != nil {
    //		return err
    //	}
    w.WriteHeader(http.StatusCreated)
    writeJSON(w, []byte("{}"))
    return nil
}

func deleteProcess(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    name := vars["name"]

    err := srv.Runtime.Processes.DeleteProcess(name)
    if err != nil {
        return err
    }
    w.WriteHeader(http.StatusOK)
    writeJSON(w, []byte("{}"))
    return nil
}

func getContainers(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    if err := parseForm(r); err != nil {
        return err
    }

    outs, err := srv.Runtime.Containers.List()
    if err != nil {
        return err
    }

    b, err := json.Marshal(outs)
    if err != nil {
        return err
    }

    writeJSON(w, b)
    return nil
}

func getContainerInfo(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    if err := parseForm(r); err != nil {
        return err
    }

    name := vars["name"]

    outs, err := srv.Runtime.Containers.GetContainerInfo(name)
    if err != nil {
        return err
    }

    b, err := json.Marshal(outs)
    if err != nil {
        return err
    }

    writeJSON(w, b)
    return nil
}

func createContainer(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    config := &ContainerInfo{}

    if err := json.NewDecoder(r.Body).Decode(config); err != nil {
        return err
    }

    err := srv.Runtime.Containers.CreateContainer(config)
    if err != nil {
        return err
    }

    w.WriteHeader(http.StatusCreated)
    writeJSON(w, []byte("{}"))
    return nil
}

func startContainer(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    //	config := &ContainerInfo{}
    //	if err := json.NewDecoder(r.Body).Decode(config); err != nil {
    //		return err
    //	}

    name := vars["name"]

    err := srv.Runtime.Containers.StartContainer(name)
    if err != nil {
        return err
    }

    w.WriteHeader(http.StatusCreated)
    writeJSON(w, []byte("{}"))
    return nil
}

func stopContainer(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    //	config := &ContainerInfo{}
    //	if err := json.NewDecoder(r.Body).Decode(config); err != nil {
    //		return err
    //	}

    name := vars["name"]

    err := srv.Runtime.Containers.StopContainer(name)
    if err != nil {
        return err
    }

    w.WriteHeader(http.StatusOK)
    writeJSON(w, []byte("{}"))
    return nil
}

func deleteContainer(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    name := vars["name"]

    err := srv.Runtime.Containers.DeleteContainer(name)
    if err != nil {
        return err
    }
    w.WriteHeader(http.StatusOK)
    writeJSON(w, []byte("{}"))
    return nil
}

func optionsHandler(srv *Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
    w.WriteHeader(http.StatusOK)
    return nil
}

func createRouter(srv *Server, logging bool) (*mux.Router, error) {
    r := mux.NewRouter()

    m := map[string]map[string]func(*Server, http.ResponseWriter, *http.Request, map[string]string) error{
        "GET": {
            "/processes":            getProcesses,
            "/containers":           getContainers,
            "/containers/{name:.*}": getContainerInfo,
        },
        "POST": {
            // TODO: processes post by name is inconsistent
            "/processes/{name:.*}": postProcess,

            "/containers":                 createContainer,
            "/containers/{name:.*}/start": startContainer,
            "/containers/{name:.*}/stop":  stopContainer,
        },
        "DELETE": {
            "/processes/{name:.*}":  deleteProcess,
            "/containers/{name:.*}": deleteContainer,
        },
        "OPTIONS": {
            "": optionsHandler,
        },
    }

    for method, routes := range m {
        for route, fct := range routes {
            Debugf("Registering %s, %s", method, route)
            // NOTE: scope issue, make sure the variables are local and won't be changed
            localRoute := route
            localMethod := method
            localFct := fct
            f := func(w http.ResponseWriter, r *http.Request) {
                Debugf("Calling %s %s", localMethod, localRoute)

                if logging {
                    log.Println(r.Method, r.RequestURI)
                }

                if err := localFct(srv, w, r, mux.Vars(r)); err != nil {
                    if logging {
                        log.Printf("Error in %s %s: %v", r.Method, r.RequestURI, err)
                    }

                    httpError(w, err)
                }
            }

            if localRoute == "" {
                r.Methods(localMethod).HandlerFunc(f)
            } else {
                //				r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(f)
                r.Path(localRoute).Methods(localMethod).HandlerFunc(f)
            }
        }
    }
    return r, nil
}

func ListenAndServe(proto, addr string, srv *Server, logging bool) error {
    log.Printf("Listening for HTTP on %s (%s)\n", addr, proto)

    r, err := createRouter(srv, logging)
    if err != nil {
        return err
    }
    l, e := net.Listen(proto, addr)
    if e != nil {
        return e
    }
    //as the daemon is launched as root, change to permission of the socket to allow non-root to connect
    if proto == "unix" {
        os.Chmod(addr, 0777)
    }
    httpSrv := http.Server{Addr: addr, Handler: r}
    return httpSrv.Serve(l)
}
