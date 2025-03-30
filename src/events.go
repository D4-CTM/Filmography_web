package renderer

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"
    _ "github.com/godror/godror"
)

//(description= (retry_count=20)(retry_delay=3)(address=(protocol=tcps)(port=1522)(host=adb.mx-queretaro-1.oraclecloud.com))(connect_data=(service_name=g32efab5c690c1e_lczydxip7wj7saab_high.adb.oraclecloud.com))(security=(ssl_server_dn_match=yes)))
//user="admin" password="Delcid_Bueso_6004" connectionString="lczydxip7wj7saab_high"
//ADMIN:Delcid_Bueso_6004@adb.mx-queretaro-1.oraclecloud.com:1522/lczydxip7wj7saab_high
func getConnetion() (*sqlx.DB, error) {
    con, err := sqlx.Open("godror", `admin:Delcid_Bueso_6004@adb.mx-queretaro-1.oraclecloud.com:1522/LCZYDXIP7WJ7SAAB_high`)
    if err != nil {
        return nil, fmt.Errorf("Crash while stablishing conection!\nerr.Error(): %v\n", err.Error())
    }
    return con, nil
}

func writeStatusMessage(w http.ResponseWriter, status int, message string) {
    w.Header().Set("HX-Status", fmt.Sprint(status))
    w.Header().Set("HX-Message", message)
    w.WriteHeader(status)
}

func EventRegisterUser(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Beginning user registration...")

    con, err := getConnetion()
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotAcceptable); 
        fmt.Println(err.Error())
        return
    }
    defer con.Close()

    err = con.Ping()
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotAcceptable); 
        fmt.Println(err.Error())
        return
    }
    fmt.Println("User registered succesfully!")
    writeStatusMessage(w, http.StatusOK, "User succesfully registered!")
}

