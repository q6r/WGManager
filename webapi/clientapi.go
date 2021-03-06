package webapi

import (
	"WGManager/webapi/resource"
	"WGManager/wg"
	"bytes"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func checkIPAccess(clientip string, allowedIPScidr []string) bool {
	ip := net.ParseIP(clientip)
	for _, aips := range allowedIPScidr {
		_, ipnet, err := net.ParseCIDR(aips)
		if err != nil {
			log.Println(err)
			return false
		}
		if ipnet.Contains(ip) {
			return true
		}
	}

	return false
}

//StartAdminClient start the REST API Echo Server for inserting watermark
func StartClient(wgConfig *wg.WGConfig) error {
	e := echo.New()
	const subserviceIdentifier = "StartWebClient"
	configureClientWebServer(e)
	configureAllRoutesClient(e, wgConfig)
	address := (wgConfig.APIListenAddress + ":" + strconv.Itoa(int(wgConfig.APIListenPort)))
	if wgConfig.APIUseTLS {
		err := e.StartTLS(address, (wgConfig.APITLSCert), (wgConfig.APITLSKey))
		if err != nil {
			panic(err)
		}
	} else {
		err := e.Start(address)
		if err != nil {
			panic(err)
		}
	}
	//

	return nil
}
func configureClientWebServer(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("100M"))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
	}))
	mime.AddExtensionType(".js", "application/javascript") //This will solve some windows shit issue, when it will serve javascript file as text/plain, read more about it at:https://github.com/labstack/echo/issues/1038

}

func configureAllRoutesClient(e *echo.Echo, wgConfig *wg.WGConfig) {
	postAllocateClient(e, wgConfig)
	postRevokeClient(e, wgConfig)
}

/*



 */
func postAllocateClient(e *echo.Echo, wgConfig *wg.WGConfig) {
	e.POST("/api/client", func(c echo.Context) error {
		u := new(resource.WgAllocateClientRequest)
		IsAllowed := checkIPAccess(c.RealIP(), wgConfig.APIAllowedIPSCIDR)
		if !IsAllowed {
			return c.String(http.StatusUnauthorized, fmt.Sprintf("You are not allowed to access, ip: %s", c.RealIP()))
		}
		if err := c.Bind(u); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		qrbytes, err := wgConfig.AllocateClient(u.Instancename, u.Clientuuid)
		responseObj := "Allocation Successfull"
		if err != nil {
			responseObj = err.Error()
			return c.JSONPretty(http.StatusBadRequest, responseObj, "  ")
		}
		return c.Stream(http.StatusOK, "image/png", bytes.NewReader(qrbytes))
		//return c.JSONPretty(http.StatusOK, responseObj, "  ")
	})
}

func postRevokeClient(e *echo.Echo, wgConfig *wg.WGConfig) {
	e.DELETE("/api/client", func(c echo.Context) error {
		u := new(resource.WgRevokeClientRequest)
		IsAllowed := checkIPAccess(c.RealIP(), wgConfig.APIAllowedIPSCIDR)
		if !IsAllowed {
			return c.String(http.StatusUnauthorized, fmt.Sprintf("You are not allowed to access, ip: %s", c.RealIP()))
		}
		if err := c.Bind(u); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		err := wgConfig.RevokeClient(u.Instancename, u.Clientuuid)
		responseObj := "Revocation Successfull"
		if err != nil {
			responseObj = err.Error()
			return c.JSONPretty(http.StatusBadRequest, responseObj, "  ")

		}

		return c.JSONPretty(http.StatusOK, responseObj, "  ")
	})
}
