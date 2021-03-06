/*
 * Copyright The Dragonfly Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package uploader

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/dragonflyoss/Dragonfly/dfget/config"
	"github.com/dragonflyoss/Dragonfly/version"

	"github.com/go-check/check"
	"github.com/gorilla/mux"
)

func init() {
	check.Suite(&UploaderUtilTestSuite{})
}

type UploaderUtilTestSuite struct {
	workHome string
	host     string
	ip       string
	port     int
	ln       net.Listener
}

func (s *UploaderUtilTestSuite) SetUpSuite(c *check.C) {
	s.workHome, _ = ioutil.TempDir("/tmp", "dfget-UploaderUtilTestSuite-")
	s.startTestServer()
}

func (s *UploaderUtilTestSuite) TearDownSuite(c *check.C) {
	s.ln.Close()
	if s.workHome != "" {
		if err := os.RemoveAll(s.workHome); err != nil {
			fmt.Printf("remove path:%s error", s.workHome)
		}
	}
}

func (s *UploaderUtilTestSuite) TestGeneratePort(c *check.C) {
	port := generatePort(0)
	c.Assert(port >= config.ServerPortLowerLimit, check.Equals, true)
	c.Assert(port <= config.ServerPortUpperLimit, check.Equals, true)
}

func (s *UploaderUtilTestSuite) TestGetPort(c *check.C) {
	metaPath := path.Join(s.workHome, "meta")
	port := getPortFromMeta(metaPath)
	c.Assert(port, check.Equals, 0)

	servicePort := 8080
	meta := config.NewMetaData(metaPath)
	meta.ServicePort = servicePort
	err := meta.Persist()
	c.Check(err, check.IsNil)

	port = getPortFromMeta(metaPath)
	c.Assert(port, check.Equals, servicePort)
}

func (s *UploaderUtilTestSuite) TestCheckServer(c *check.C) {
	// normal test
	result, err := checkServer(s.ip, s.port, s.workHome, commonFile, 0)
	c.Check(err, check.IsNil)
	c.Check(result, check.Equals, commonFile)

	// error url test
	result, err = checkServer(s.ip+"1", s.port, s.workHome, commonFile, 0)
	c.Check(err, check.NotNil)
	c.Check(result, check.Equals, "")
}

func (s *UploaderUtilTestSuite) TestUpdateServicePortInMeta(c *check.C) {
	expectedPort := 80
	metaPath := path.Join(s.workHome, "meta")
	updateServicePortInMeta(metaPath, expectedPort)
	port := getPortFromMeta(metaPath)
	c.Assert(port, check.Equals, expectedPort)
}

func (s *UploaderUtilTestSuite) startTestServer() {
	// run a server
	s.ip = "127.0.0.1"
	s.port = rand.Intn(1000) + 63000
	s.host = fmt.Sprintf("%s:%d", s.ip, s.port)
	s.ln, _ = net.Listen("tcp", s.host)
	checkHandler := func(w http.ResponseWriter, r *http.Request) {
		fileName := mux.Vars(r)["commonFile"]
		fmt.Fprintf(w, "%s@%s", fileName, version.DFGetVersion)
	}
	r := mux.NewRouter()
	r.HandleFunc(config.LocalHTTPPathCheck+"{commonFile:.*}", checkHandler).Methods("GET")
	go http.Serve(s.ln, r)
}
