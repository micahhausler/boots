// TODO this is mostly copied from tinkerbell/tink.
// Need to abstract/externalize it
package packet

import (
	"context"
	"fmt"

	"github.com/tinkerbell/boots/k8s/api/v1alpha1"
	"github.com/tinkerbell/boots/pkg/controllers"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"

	tw "github.com/tinkerbell/tink/protos/workflow"
	"k8s.io/apimachinery/pkg/types"
)

func K8sActionListToTink(wf *v1alpha1.Workflow) *tw.WorkflowActionList {
	resp := &tw.WorkflowActionList{
		ActionList: []*tw.WorkflowAction{},
	}
	for _, task := range wf.Status.Tasks {
		for _, action := range task.Actions {
			resp.ActionList = append(resp.ActionList, &tw.WorkflowAction{
				TaskName: task.Name,
				Name:     action.Name,
				Image:    action.Image,
				Timeout:  action.Timeout,
				Command:  action.Command,
				WorkerId: task.WorkerAddr,
				Volumes:  task.Volumes,
				Environment: func(env map[string]string) []string {
					resp := []string{}
					merged := map[string]string{}
					for k, v := range env {
						merged[k] = v
					}
					for k, v := range action.Environment {
						merged[k] = v
					}
					for k, v := range merged {
						resp = append(resp, fmt.Sprintf("%s=%s", k, v))
					}
					return resp
				}(task.Environment),
				// Pid: action.Pid,
			})
		}
	}
	return resp
}

func (c *client) getWorkflowsForWorker(ctx context.Context, id string) ([]string, error) {
	workflowList := &v1alpha1.WorkflowList{}
	err := c.manager.GetClient().List(ctx, workflowList, &crclient.MatchingFields{
		controllers.WorkflowWorkerKey: id,
	})
	if err != nil {
		return nil, err
	}
	wfIds := []string{}
	for _, wf := range workflowList.Items {
		wfIds = append(wfIds, wf.Name)
	}
	return wfIds, nil
}

// GetWorkflowContexts implements tinkerbell.GetWorkflowContexts
func (c *client) GetWorkflowContextList(ctx context.Context, workerId string) (*tw.WorkflowContextList, error) {
	wfs, err := c.getWorkflowsForWorker(ctx, workerId)
	if err != nil {
		return nil, err
	}
	wfContexts := []*tw.WorkflowContext{}
	for _, wf := range wfs {

		wfContext, err := c.getWorkflowContexts(context.Background(), wf)
		if err != nil {
			return nil, err
		}
		if c.isApplicableToSend(context.Background(), wfContext, workerId) {
			wfContexts = append(wfContexts, wfContext)
		}
	}
	return &tw.WorkflowContextList{
		WorkflowContexts: wfContexts,
	}, nil
}

// Called by worker
func (c *client) getWorkflowActions(ctx context.Context, wfID string) (*tw.WorkflowActionList, error) {
	wf, err := c.getWorkflowByName(ctx, wfID)
	if err != nil {
		return nil, err
	}
	return K8sActionListToTink(wf), nil
}

// isApplicableToSend checks if a particular workflow context is applicable or if it is needed to
// be sent to a worker based on the state of the current action and the targeted workerID
func (c *client) isApplicableToSend(context context.Context, wfContext *tw.WorkflowContext, workerID string) bool {
	if wfContext.GetCurrentActionState() == tw.State_STATE_FAILED ||
		wfContext.GetCurrentActionState() == tw.State_STATE_TIMEOUT {
		return false
	}
	actions, err := c.getWorkflowActions(context, wfContext.GetWorkflowId())
	if err != nil {
		return false
	}
	if wfContext.GetCurrentActionState() == tw.State_STATE_SUCCESS {
		if isLastAction(wfContext, actions) {
			return false
		}
		if wfContext.GetCurrentActionIndex() == 0 {
			if actions.ActionList[wfContext.GetCurrentActionIndex()+1].GetWorkerId() == workerID {
				return true
			}
		}
	} else if actions.ActionList[wfContext.GetCurrentActionIndex()].GetWorkerId() == workerID {
		return true

	}
	return false
}

func isLastAction(wfContext *tw.WorkflowContext, actions *tw.WorkflowActionList) bool {
	return int(wfContext.GetCurrentActionIndex()) == len(actions.GetActionList())-1
}

func (c *client) getWorkflowByName(ctx context.Context, name string) (*v1alpha1.Workflow, error) {
	workflow := &v1alpha1.Workflow{}
	err := c.manager.GetClient().Get(ctx, types.NamespacedName{Name: name}, workflow)
	if err != nil {
		return nil, err
	}
	return workflow, nil
}

func (c *client) getWorkflowContexts(ctx context.Context, wfID string) (*tw.WorkflowContext, error) {
	wf, err := c.getWorkflowByName(ctx, wfID)
	if err != nil {
		return nil, err
	}

	var (
		found           bool
		taskIndex       int
		taskActionIndex int
		actionIndex     int
		actionCount     int
	)
	for ti, task := range wf.Status.Tasks {
		for ai, action := range task.Actions {
			actionCount++
			if (action.Status == tw.State_name[int32(tw.State_STATE_PENDING)] || action.Status == tw.State_name[int32(tw.State_STATE_RUNNING)]) && !found {
				taskIndex = ti
				actionIndex = ai
				found = true
			}
			if !found {
				actionIndex++
			}
		}
	}

	resp := &tw.WorkflowContext{
		WorkflowId:           wfID,
		CurrentWorker:        wf.Status.Tasks[taskIndex].WorkerAddr,
		CurrentTask:          wf.Status.Tasks[taskIndex].Name,
		CurrentAction:        wf.Status.Tasks[taskIndex].Actions[taskActionIndex].Name,
		CurrentActionIndex:   int64(actionIndex),
		CurrentActionState:   tw.State(tw.State_value[wf.Status.Tasks[taskIndex].Actions[taskActionIndex].Status]),
		TotalNumberOfActions: int64(actionCount),
	}
	return resp, nil
}
