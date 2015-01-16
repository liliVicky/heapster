// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/heapster/sinks"
	"github.com/GoogleCloudPlatform/heapster/sources"
	"github.com/GoogleCloudPlatform/heapster/validate"
	"github.com/GoogleCloudPlatform/heapster/version"
	"github.com/golang/glog"
)

var argPollDuration = flag.Duration("poll_duration", 10*time.Second, "Polling duration")
var argPort = flag.Int("port", 8082, "port to listen")
var argIp = flag.String("listen_ip", "", "IP to listen on, defaults to all IPs")

func main() {
	flag.Parse()
	glog.Infof(strings.Join(os.Args, " "))
	glog.Infof("Heapster version %v", version.HeapsterVersion)
	err := doWork()
	if err != nil {
		glog.Error(err)
		os.Exit(1)
	}
	setupHandlers()
	addr := fmt.Sprintf("%s:%d", *argIp, *argPort)
	glog.Infof("Starting heapster on port %d", *argPort)
	glog.Fatal(http.ListenAndServe(addr, nil))
	os.Exit(0)
}

func setupHandlers() {
	// Validation/Debug handler.
	http.HandleFunc(validate.ValidatePage, func(w http.ResponseWriter, r *http.Request) {
		err := validate.HandleRequest(w)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
		}
	})

	// TODO(jnagal): Add a main status page.
	http.Handle("/", http.RedirectHandler(validate.ValidatePage, http.StatusTemporaryRedirect))
}

func doWork() error {
	source, err := sources.NewSource()
	if err != nil {
		return err
	}
	sink, err := sinks.NewSink()
	if err != nil {
		return err
	}
	go housekeep(source, sink)
	return nil
}

func housekeep(source sources.Source, sink sinks.Sink) {
	ticker := time.NewTicker(*argPollDuration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			data, err := source.GetInfo()
			if err != nil {
				glog.Fatalf("failed to get information from source")
			}
			if err := sink.StoreData(data); err != nil {
				glog.Fatalf("failed to push information to sink")

			}
		}
	}

}
