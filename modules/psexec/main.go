package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/base64"
	"strings"
)

type Task struct {
	TaskId      string   `json:"taskid"`
	ServiceName string   `json:"servicename"`
	ServiceDesc string   `json:"servicedesc"`
	BinPath     string   `json:"binpath"`
	Hostname    string   `json:"hostname"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Domain      string   `json:"domain"`
	FileData    []byte   `json:"filedata"`
	Args        []string `json:"args"`

	// General Purpose
	status   string
	tasktype string
}

func InitialiseTask(task *Task) error {
	task.tasktype = "resppsexec"
	if task.Args == nil {
		task.Args = []string{""}
	}
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
	password := task.Password
	if password != "" && password != "_" {
		decoded, err := base64.StdEncoding.DecodeString(password)
		if err != nil {
			Println("Failed to decode password: ", err.Error())
			task.status = "failed"
			return []byte("Failed to decode password: " + err.Error()), err
		}
		password = string(decoded)
	}

	err = PsExec(task.Hostname, task.BinPath, task.FileData, strings.Join(task.Args, " "), task.ServiceName, task.ServiceDesc, task.Username, password, task.Domain)
	if err != nil {
		Println("Error Occured --> ", err.Error())
		task.status = "failed"
		return []byte(err.Error()), err
	}
	task.status = "completed"
	return []byte("PsExec returned without errors"), err
}

func IsTaskDone(task Task) bool {
	if task.status == "completed" || task.status == "failed" {
		return true
	}
	return false
}

func ComputeTaskStatus(task Task) string {
	return task.status
}

func ComputeTaskType(task Task) string {
	return task.tasktype
}
