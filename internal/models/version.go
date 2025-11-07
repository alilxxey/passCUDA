package models

import "encoding/json"

type VersionResponseModel struct {
	Name         string `json:"name"`
	GitCommit    string `json:"git_commit"`
	GitBranch    string `json:"git_branch"`
	GitBranchNum string `json:"git_branch_num"`
	BuildDate    string `json:"build_date"`
	BuildTime    string `json:"build_time"`
	Version      string `json:"version"`
}

// compile-time variables
var gitCommit string
var gitBranch string
var gitBranchNum string
var buildDate string
var buildTime string
var version string

var versionResponseData = &VersionResponseModel{
	Name:         "passCUDA",
	GitCommit:    gitCommit,
	GitBranch:    gitBranch,
	GitBranchNum: gitBranchNum,
	BuildDate:    buildDate,
	BuildTime:    buildTime,
	Version:      version,
}

func GetPrintableVersionInfo() (string, error) {
	jsonData, err := json.MarshalIndent(versionResponseData, "", "    ")

	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func GetVersionResponseModel() *VersionResponseModel {
	return versionResponseData
}
