/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
)

const pods = 100

func launchPods() {
	path := "sample_podspec_template.yaml"
	t, err := template.ParseFiles(path)
	if err != nil {
		log.Print(err)
		return
	}

	for i := 0; i < pods; i++ {
		generatedFilename := fmt.Sprintf("generated/hello_%d.yaml", i)
		createFileUsingTemplate(t, generatedFilename, &struct{ Count int }{Count: i})
		// read template file
		// generate resolved
		// run kubectl apply dir
	}
}

func createFileUsingTemplate(t *template.Template, filename string, data interface{}) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = t.Execute(f, data)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	launchPods()
}
