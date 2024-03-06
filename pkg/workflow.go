package pkg

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Job struct {
	Name     string `yaml:"name"`
	Fastened bool   `yaml:"fastened"`
}

type Workflow struct {
	FileName string
	Name     string
	Jobs     []*Job
}

type WorkflowClient struct {
	WorkflowDir string
}

func NewWorkflowClient(workflowDir string) *WorkflowClient {
	return &WorkflowClient{WorkflowDir: workflowDir}
}

func (w *WorkflowClient) GetWorkflows() ([]*Workflow, error) {
	workflowDir := w.WorkflowDir
	workflowFiles, err := os.ReadDir(workflowDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("no workflows found in %s", workflowDir)
	} else if err != nil {
		return nil, err
	}

	workflows := []*Workflow{}

	for _, workflowFile := range workflowFiles {
		workflowFileName := workflowFile.Name()
		isYAML := strings.HasSuffix(workflowFileName, ".yml") || strings.HasSuffix(workflowFileName, ".yaml")
		if !workflowFile.IsDir() && isYAML {
			workflowContent, err := os.ReadFile(fmt.Sprintf("%s/%s", workflowDir, workflowFileName))
			if err != nil {
				return nil, err
			}
			workflow := map[string]interface{}{}
			err = yaml.Unmarshal(workflowContent, &workflow)
			if err != nil {
				return nil, err
			}
			jobs := []*Job{}
			for jobName, job := range workflow["jobs"].(map[interface{}]interface{}) {
				jobMap := job.(map[interface{}]interface{})
				fastened := false
				for _, step := range jobMap["steps"].([]interface{}) {
					stepMap := step.(map[interface{}]interface{})
					if stepMap["uses"] != nil {
						uses := stepMap["uses"].(string)
						if strings.HasPrefix(uses, "koalalab-inc/bolt@") {
							fastened = true
							break
						}
					}
				}
				jobs = append(jobs, &Job{Name: jobName.(string), Fastened: fastened})
			}
			workflowName := ""
			if workflow["name"] != nil {
				workflowName = workflow["name"].(string)
			}
			workflows = append(workflows, &Workflow{Name: workflowName, FileName: workflowFileName, Jobs: jobs})
		}
	}
	return workflows, nil
}

func (w *WorkflowClient) AddBoltToWorkflow(workflowFileName string, jobsToBeBolted []string) error {
	workflowDir := w.WorkflowDir
	content, err := os.ReadFile(fmt.Sprintf("%s/%s", workflowDir, workflowFileName))
	if err != nil {
		return err
	}
	workflow := map[string]interface{}{}
	err = yaml.Unmarshal(content, &workflow)
	if err != nil {
		return err
	}
	boltedJobs := []string{}

	for jobName, job := range workflow["jobs"].(map[interface{}]interface{}) {
		jobMap := job.(map[interface{}]interface{})
		for _, step := range jobMap["steps"].([]interface{}) {
			stepMap := step.(map[interface{}]interface{})
			if stepMap["uses"] != nil {
				uses := stepMap["uses"].(string)
				if strings.HasPrefix(uses, "koalalab-inc/bolt@") {
					boltedJobs = append(boltedJobs, jobName.(string))
				}
			}
		}
	}

	lines := strings.Split(string(content), "\n")
	linesCount := len(lines)
	newLines := []string{}
	embedBolt := false
	indentation := strings.Repeat(" ", 2)
	//steps field will always have 2 indetations
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "steps:") {
			if stepsIndex := strings.Index(line, "steps:"); stepsIndex != -1 {
				indentationLength := (stepsIndex + 1) / 2
				indentation = line[0:indentationLength]
			}
		}
	}
	jobsSection := false
	var alreadyBolted bool
	var jobSelected bool
	for index, line := range lines {
		if strings.HasPrefix(line, "jobs:") || strings.HasPrefix(line, "jobs ") {
			jobsSection = true
		} else if jobsSection {
			isComment := strings.HasPrefix(strings.TrimSpace(line), "#")
			hasIndentation := strings.HasPrefix(line, indentation)
			if len(strings.TrimSpace(line)) > 0 && !isComment && !hasIndentation {
				jobsSection = false
			}
		}
		if jobsSection {
			hasIndentation := strings.HasPrefix(line, indentation)
			hasDoubleIndentation := strings.HasPrefix(line, indentation+indentation)
			isComment := strings.HasPrefix(strings.TrimSpace(line), "#")

			if hasIndentation && !hasDoubleIndentation && !isComment {
				alreadyBolted = false
				for _, boltedJob := range boltedJobs {
					alreadyBolted = strings.HasPrefix(strings.TrimSpace(line), boltedJob+":") || strings.HasPrefix(strings.TrimSpace(line), boltedJob+" ")
					if alreadyBolted {
						break
					}
				}
				jobSelected = false
				for _, jobToBeBolted := range jobsToBeBolted {
					jobSelected = strings.HasPrefix(strings.TrimSpace(line), jobToBeBolted+":") || strings.HasPrefix(strings.TrimSpace(line), jobToBeBolted+" ")
					if jobSelected {
						break
					}
				}
			}
			if embedBolt && !alreadyBolted && jobSelected {
				whitespaces := strings.Repeat(" ", 6)
				for i := index; i < linesCount; i++ {
					trimmedLine := strings.TrimSpace(lines[i])
					if strings.HasPrefix(trimmedLine, "-") {
						whitespaces = strings.Split(lines[i], "-")[0]
						break
					}
				}
				newLines = append(newLines, fmt.Sprintf("%s- name: Setup Bolt", whitespaces))
				newLines = append(newLines, fmt.Sprintf("%s%suses: koalalab-inc/bolt@v1", whitespaces, indentation))
				embedBolt = false
			} else if alreadyBolted {
				embedBolt = false
			}
			if strings.HasPrefix(strings.TrimSpace(line), "steps:") || strings.HasPrefix(strings.TrimSpace(line), "steps ") {
				if stepsIndex := strings.Index(line, "steps:"); stepsIndex != -1 {
					embedBolt = true
				} else if stepsIndex := strings.Index(line, "steps "); stepsIndex != -1 {
					embedBolt = true
				}
			}
		}
		newLines = append(newLines, line)
	}

	modifiedContent := strings.Join(newLines, "\n")

	err = os.WriteFile(fmt.Sprintf("%s/%s", workflowDir, workflowFileName), []byte(modifiedContent), 0644)

	if err != nil {
		return err
	}

	return nil
}
