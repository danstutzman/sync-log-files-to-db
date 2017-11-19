package main

type Table struct {
	Rows []map[string]interface{}
}

type Dataset struct {
	Tables map[string]Table
}

type Project struct {
	Datasets map[string]Dataset
}

var projects = map[string]Project{}

var queryByJobId = map[string]string{}
