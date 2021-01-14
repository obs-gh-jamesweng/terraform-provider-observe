package meta

import (
	"errors"
	"fmt"
	"time"
)

type Workspace struct {
	ID       ObjectIdScalar `json:"id"`
	Label    string         `json:"label"`
	Datasets []*Dataset     `json:"datasets"`
}

type DatasetSaveResult struct {
	Dataset *Dataset `json:"dataset"`
}

type Dataset struct {
	ID               ObjectIdScalar `json:"id"`
	WorkspaceId      ObjectIdScalar `json:"workspaceId"`
	Version          string         `json:"version"`
	Label            string         `json:"label"`
	LatencyDesired   *time.Duration `json:"latencyDesired"`
	FreshnessDesired *time.Duration `json:"freshnessDesired"`
	Typedef          Typedef        `json:"typedef"`
	Description      *string        `json:"description"`
	IconURL          *string        `json:"iconUrl"`
	PathCost         *int64         `json:"pathCost"`
	Transform        *Transform     `json:"transform"`
	ForeignKeys      []ForeignKey   `json:"foreignKeys"`
}

func (d *Dataset) Decode(v interface{}) error {
	return decodeStrict(v, d)
}

type Typedef struct {
	Definition map[string]interface{} `json:"definition"`
}

type Transform struct {
	Dataset *Dataset          `json:"dataset"`
	ID      ObjectIdScalar    `json:"id"`
	Current *TransformVersion `json:"current"`
}

type TransformVersion struct {
	Transform *Transform       `json:"transform"`
	Query     *MultiStageQuery `json:"query"`
}

type MultiStageQuery struct {
	OutputStage string        `json:"outputStage"`
	Stages      []*StageQuery `json:"stages"`
}

type StageQuery struct {
	ID       string             `json:"id"`
	Input    []*InputDefinition `json:"input"`
	Pipeline string             `json:"pipeline"`
}

type InputDefinition struct {
	InputName   string          `json:"inputName"`
	InputRole   *InputRole      `json:"inputRole"`
	DatasetID   *ObjectIdScalar `json:"datasetId,omitempty"`
	DatasetPath *string         `json:"datasetPath"`
	StageID     string          `json:"stageId"`
}

type InputRole string

const (
	InputRoleDefault   InputRole = ""
	InputRoleData      InputRole = "Data"
	InputRoleReference InputRole = "Reference"
)

func (e InputRole) IsValid() bool {
	switch e {
	case InputRoleDefault,
		InputRoleData,
		InputRoleReference:
		return true
	}
	return false
}

func (e InputRole) String() string {
	return string(e)
}

type ResultStatus struct {
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"errorMessage"`
	DetailedInfo map[string]interface{} `json:"detailedInfo"`
}

func (s *ResultStatus) Error() error {
	if s.Success {
		return nil
	}
	if s.ErrorMessage != "" {
		return fmt.Errorf("request failed: %q", s.ErrorMessage)
	}
	return errors.New("request failed")
}

type ForeignKey struct {
	TargetDataset        *int64   `json:"targetDataset,string"`
	TargetStageLabel     *string  `json:"targetStageLabel"`
	Label                *string  `json:"label"`
	TargetLabelFieldName *string  `json:"targetLabelFieldName"`
	SrcFields            []string `json:"srcFields"`
	DstFields            []string `json:"dstFields"`
}

type DeferredForeignKey struct {
	ID            ObjectIdScalar           `json:"id"`
	WorkspaceID   ObjectIdScalar           `json:"workspaceId"`
	SourceDataset DeferredDatasetReference `json:"source"`
	TargetDataset DeferredDatasetReference `json:"target"`
	SrcFields     []string                 `json:"srcFields"`
	DstFields     []string                 `json:"dstFields"`
	Label         *string                  `json:"label,omitempty"`
	Resolution    *ResolvedForeignKey      `json:"resolution,omitempty"`
	Status        DeferredForeignKeyStatus `json:"status"`
}

type ResolvedForeignKey struct {
	SourceID ObjectIdScalar `json:"sourceId"`
	TargetID ObjectIdScalar `json:"targetId"`
}

//  If the foreign key doesn't match some datasets, that may be because of a
//  number of reasons. A "true" in a status field means that that part of the
//  resolution is A-OK, a "false" means that an error happened in that part.
type DeferredForeignKeyStatus struct {
	ID                      ObjectIdScalar `json:"id"`
	FoundSource             bool           `json:"foundSource"`
	FoundTarget             bool           `json:"foundTarget"`
	MatchedSourceFields     bool           `json:"matchedSourceFields"`
	MatchedTargetFields     bool           `json:"matchedTargetFields"`
	FieldTypesAreComparable bool           `json:"fieldTypesAreComparable"`
	//  EnglishError is empty if there's no error, else it's a human-readable
	//  string describing what things prevent the key from resolving.
	ErrorText string `json:"errorText"`
}

type DeferredDatasetReference struct {
	DatasetID   *ObjectIdScalar `json:"datasetId,omitempty"`
	DatasetPath *string         `json:"datasetPath,omitempty"`
}

type BookmarkGroup struct {
	ID           ObjectIdScalar            `json:"id"`
	Name         string                    `json:"name"`
	IconURL      string                    `json:"iconUrl"`
	WorkspaceID  ObjectIdScalar            `json:"workspaceId"`
	Presentation BookmarkGroupPresentation `json:"presentation"`
}

type Bookmark struct {
	ID       ObjectIdScalar `json:"id"`
	Name     string         `json:"name"`
	IconURL  string         `json:"iconUrl"`
	TargetID ObjectIdScalar `json:"targetId"`
	// TODO: use ObjectKind
	TargetIDKind string         `json:"targetIdKind"`
	GroupID      ObjectIdScalar `json:"groupId"`
}

type ChannelAction struct {
	ID          ObjectIdScalar `json:"id"`
	Name        string         `json:"name"`
	IconURL     *string        `json:"iconUrl"`
	Description *string        `json:"description"`
	WorkspaceId ObjectIdScalar `json:"workspaceId"`
	//CreatedBy   UserIdScalar   `json:"createdBy"`
	//CreatedDate TimeScalar     `json:"createdDate"`
	//UpdatedBy   UserIdScalar   `json:"updatedBy"`
	//UpdatedDate TimeScalar     `json:"updatedDate"`
	//RateLimit   *time.Duration `json:"rateLimit"`

	Webhook *WebhookChannelAction `json:"webhook"`
	Email   *EmailChannelAction   `json:"email"`
}

type WebhookChannelAction struct {
	URLTemplate  *string          `json:"urlTemplate"`
	Method       *string          `json:"method"`
	BodyTemplate *string          `json:"bodyTemplate"`
	Headers      []*WebhookHeader `json:"headers"`
}

type EmailChannelAction struct {
	TargetAddresses []string `json:"targetAddresses"`
	SubjectTemplate *string  `json:"subjectTemplate"`
	BodyTemplate    *string  `json:"bodyTemplate"`
	IsHTML          bool     `json:"isHtml"`
}

type Channel struct {
	ID          ObjectIdScalar `json:"id"`
	Name        string         `json:"name"`
	IconURL     *string        `json:"iconUrl"`
	Description *string        `json:"description"`
	WorkspaceId ObjectIdScalar `json:"workspaceId"`
	// in theory, we could use *ChannelAction, but unmarshalling is a pain
	// because of subtypes
	Actions []struct {
		ID ObjectIdScalar `json:"id"`
	} `json:"actions"`
}

type TaskResult struct {
	QueryID string `json:"queryId"`
	StageID string `json:"stageId"`
	// The time range which this set of results cover.
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	//ResultProgress   *TaskResultProgress
	ResultCursor *SnowflakeCursor `json:"resultCursor"`
	//PaginatedResults *metatypes.PaginatedResults
	ResultKind   *ResultKind       `json:"resultKind"`
	ResultSchema *TaskResultSchema `json:"resultSchema"`
	//ParsedPipeline   *metatypes.ParsedPipeline
	Error *string `json:"error"`
	//EstimatedCost    []CostMetric
}

type TaskResultSchema struct {
	TypedefDefinition struct {
		Fields []map[string]interface{} `json:"fields"`
	} `json:"typedefDefinition"`
}

type SnowflakeCursor struct {
	QueryID       string                   `json:"queryId,omitempty"`
	TotalRowCount int64                    `json:"totalRowCount,omitempty"`
	ColumnDesc    []map[string]interface{} `json:"columnDesc,omitempty"`
	Columns       [][]*string              `json:"columns,omitempty"`
	Chunks        interface{}              `json:"chunks,omitempty"`
	HttpHeaders   map[string]string        `json:"httpHeaders,omitempty"`
}
