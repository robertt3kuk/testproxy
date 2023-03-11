package main

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var requests = make(map[string]string)

func main() {
	r := gin.Default()

	r.POST("/", Proxy)

	r.GET("/:id", func(c *gin.Context) {
		id := c.Param("id")
		body, ok := requests[id]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "request not found"})
			return
		}

		c.String(http.StatusOK, body)
	})

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type Response struct {
	ID      string              `json:"id"`
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Length  int                 `json:"length"`
}

func Proxy(c *gin.Context) {
	var req Request

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := http.Client{}
	httpReq, err := http.NewRequest(req.Method, req.URL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for key, value := range req.Headers {
		httpReq.Header.Add(key, value)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := ResponseFormation(resp)
	result.Length = len(body)

	requests[result.ID] = string(body)

	c.JSON(http.StatusOK, result)
}

func ResponseFormation(resp *http.Response) Response {
	return Response{
		ID:     uuid.NewString(),
		Status: resp.StatusCode,
		Headers: func() map[string][]string {
			m := make(map[string][]string)
			for key, value := range resp.Header {
				m[key] = value
			}
			return m
		}(),
	}
}
