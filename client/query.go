package client

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/observeinc/terraform-provider-observe/client/internal/meta"
)

type Query struct {
	Inputs map[string]*Input `json:"inputs"`
	Stages []*Stage          `json:"stages"`
}

// Stage applies a pipeline to an input
// If no input is provided, stage will follow on from previous stage
// An alias must be provided for callers to be able to reference this stage in OPAL
// Internally, the alias does not map to the stageID - it is the input name we
// use when refering to this stage
type Stage struct {
	Alias    *string `json:"alias,omitempty"`
	Input    *string `json:"input,omitempty"`
	Pipeline string  `json:"pipeline"`
}

// Input references an existing data source
type Input struct {
	Dataset *string ` json:"dataset,omitempty"`
}

func validateInput(i *Input) error {
	switch {
	case invalidObjectID(i.Dataset):
		return fmt.Errorf("dataset: %w", errObjectIDInvalid)
	case i.Dataset == nil:
		return errInputEmpty
	}
	return nil
}

func newQuery(gqlQuery *meta.MultiStageQuery) (*Query, error) {
	query := &Query{Inputs: make(map[string]*Input)}

	// first reconstruct all inputs
	stageIDs := make(map[string]string)
	for _, stageQuery := range gqlQuery.Stages {
		for _, i := range stageQuery.Input {
			if i.DatasetID != nil {
				datasetID := i.DatasetID.String()
				query.Inputs[i.InputName] = &Input{Dataset: &datasetID}
			}
			if i.StageID != "" {
				stageIDs[i.StageID] = i.InputName
			}
		}
	}

	for i, gqlStage := range gqlQuery.Stages {
		stage := &Stage{
			Pipeline: gqlStage.Pipeline,
		}

		if name, ok := stageIDs[gqlStage.StageID]; ok && name != gqlStage.StageID {
			stage.Alias = &name
		}

		inputName := gqlStage.Input[0].InputName

		switch {
		case i == 0 && len(query.Inputs) == 1:
			// defaulted to first input
		case i > 0 && query.Stages[i-1].Alias != nil && inputName == *(query.Stages[i-1].Alias):
			// follow on from aliased stage
		case stageIDs[inputName] != "":
			// follow on from anonymous stage
		default:
			stage.Input = &inputName
		}

		query.Stages = append(query.Stages, stage)
	}

	return query, nil
}

func (q *Query) toGQL() (*meta.MultiStageQueryInput, error) {
	var gqlQuery meta.MultiStageQueryInput

	// validate and convert all inputs
	var sortedNames []string
	gqlInputs := make(map[string]*meta.InputDefinitionInput, len(q.Inputs))
	for name, input := range q.Inputs {
		if err := validateInput(input); err != nil {
			return nil, fmt.Errorf("invalid input: %w", err)
		}
		gqlInputs[name] = &meta.InputDefinitionInput{
			InputName: name,
			DatasetID: toObjectPointer(input.Dataset),
		}
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	var defaultInput *meta.InputDefinitionInput
	switch len(q.Inputs) {
	case 0:
		return nil, errInputsMissing
	case 1:
		// in only one input is provided, use it as input for first stage
		defaultInput = gqlInputs[sortedNames[0]]
	}

	// We're now ready to convert stages
	// If a stage is named, it can be used as an input for every subsequent stage.
	// If a stage is anonymous, it can still be used as a default input on the next stage.
	for i, stage := range q.Stages {
		if stage.Pipeline == "" {
			return nil, fmt.Errorf("stage %d: %w", i, errStagePipelineMissing)
		}

		// Each stage will be given an ID based on the hash of all preceeding pipelines
		gqlStage := &meta.StageQueryInput{
			StageID:  fmt.Sprintf("stage-%d", i),
			Pipeline: stage.Pipeline,
		}

		// if stage has a declared input, update defaultInput
		if stage.Input != nil {
			v, ok := gqlInputs[*stage.Input]
			if !ok {
				return nil, fmt.Errorf("stage-%d: %q: %w", i, *stage.Input, errStageInputUnresolved)
			}
			defaultInput = v
		}

		if defaultInput == nil {
			return nil, fmt.Errorf("stage-%d: %w", i, errStageInputMissing)
		}

		// construct stage inputs, first default, then any declared input that
		// is referenced in pipeline.
		gqlStage.Input = append(gqlStage.Input, *defaultInput)

		for _, name := range sortedNames {
			gqlInput := gqlInputs[name]
			// don't add defaultInput a second time
			if gqlInput != defaultInput && strings.Contains(stage.Pipeline, "@"+gqlInput.InputName) {
				gqlStage.Input = append(gqlStage.Input, *gqlInput)
			}
		}

		// stage is done, append to transform
		gqlQuery.Stages = append(gqlQuery.Stages, gqlStage)
		gqlQuery.OutputStage = gqlStage.StageID

		// prepare for next iteration of loop
		// this stage will become defaultInput for the next
		defaultInput = &meta.InputDefinitionInput{
			InputName: gqlStage.StageID,
			StageID:   gqlStage.StageID,
		}

		// if explicitly named, this stage can be also be an input for the next
		if stage.Alias != nil {
			defaultInput.InputName = *stage.Alias
			// conflict?
			gqlInputs[*stage.Alias] = defaultInput
			sortedNames = append(sortedNames, *stage.Alias)
		}
	}

	// a query must have at least one stage
	if gqlQuery.OutputStage == "" {
		return nil, errStagesMissing
	}

	return &gqlQuery, nil
}

type QueryConfig struct {
	*Query
	Limit int64     `json:"limit"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (q *QueryConfig) toGQL() ([]*meta.StageInput, *meta.QueryParams, error) {
	multiStageQueryInput, err := q.Query.toGQL()
	if err != nil {
		return nil, nil, err
	}

	// This is insane. StageQueryInput is a subset of StageInput, but differs
	// in the key of the input field: one has "input", the other "inputs".
	// Convert here rather than replicating all the conversion logic.
	stages := make([]*meta.StageInput, len(multiStageQueryInput.Stages))

	var (
		resultKindData     = meta.ResultKindResultKindData
		resultKindSchema   = meta.ResultKindResultKindSchema
		resultKindSuppress = meta.ResultKindResultKindSuppress
	)

	for i, s := range multiStageQueryInput.Stages {
		stages[i] = &meta.StageInput{
			Input:    s.Input,
			StageID:  s.StageID,
			Pipeline: s.Pipeline,
			Presentation: &meta.StagePresentationInput{
				ResultKinds: []*meta.ResultKind{&resultKindSuppress},
			},
		}
	}

	outputStage := stages[len(stages)-1]
	outputStage.Presentation.ResultKinds = []*meta.ResultKind{&resultKindData, &resultKindSchema}
	outputStage.Presentation.Limit = &q.Limit

	return stages, &meta.QueryParams{
		StartTime: &q.Start,
		EndTime:   &q.End,
	}, nil
}

func newQueryResult(taskResults []*meta.TaskResult) (*QueryResult, error) {
	if len(taskResults) != 1 {
		return nil, fmt.Errorf("unexpected number of taskResults")
	}

	result := taskResults[0]

	if result.Error != nil {
		return nil, fmt.Errorf(*result.Error)
	}

	var (
		numRows = result.ResultCursor.TotalRowCount
		numCols = int64(len(result.ResultCursor.Columns))
	)

	q := &QueryResult{
		ID:        result.QueryID,
		StartTime: *result.StartTime,
		EndTime:   *result.EndTime,
		RowCount:  numRows,
		Fields:    result.ResultSchema.TypedefDefinition.Fields,
	}

	var colNames []string
	var colTypes []map[string]interface{}

	for _, f := range result.ResultSchema.TypedefDefinition.Fields {
		colNames = append(colNames, f["name"].(string))
		colTypes = append(colTypes, f["type"].(map[string]interface{}))
	}

	// convert from columnar format to list of JSONs
	// This allows output to then be parsed by command line tools such as jq
	rows := make([]map[string]interface{}, numRows)
	for i := int64(0); i < numRows; i++ {
		rows[i] = make(map[string]interface{}, numCols)

		for j := int64(0); j < numCols; j++ {
			var value interface{}
			var err error

			if cell := result.ResultCursor.Columns[j][i]; cell != nil {
				switch colTypes[j]["rep"].(string) {
				case "any", "object":
					value = json.RawMessage([]byte(*cell))
				case "duration", "int64":
					value, err = strconv.ParseInt(*cell, 10, 64)
				case "float64":
					value, err = strconv.ParseFloat(*cell, 64)
				case "timestamp":
					value, err = strconv.ParseInt(*cell, 10, 64)
					if err == nil {
						value = time.Unix(0, value.(int64)).UTC()
					}
				default:
					value = cell
				}

				if err != nil {
					return nil, fmt.Errorf("failed to cast value: %w", err)
				}
			}
			rows[i][colNames[j]] = value
		}
	}

	data, err := json.Marshal(rows)
	q.JSON = data
	return q, err
}

type QueryResult struct {
	ID        string
	StartTime time.Time
	EndTime   time.Time
	RowCount  int64
	Fields    []map[string]interface{}
	JSON      []byte
}
