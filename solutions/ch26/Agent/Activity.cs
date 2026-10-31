namespace Ensemble;

public sealed partial class Agent
{
    private partial WorkActivity ReadWorkActivity()
    {
        // Called only on the actor, including its reliable status snapshot.
        // Closing admission does not by itself finish in-flight work.
        lock (ingress)
        {
            if (transportPending || engineSends != 0)
            {
                return new WorkActivity(WorkPhase.Thinking, Detail: "Waiting for the model response");
            }
        }
        if (memoryHeadroomTaskID is not null)
        {
            return new WorkActivity(WorkPhase.Waiting, Detail: "Waiting for memory headroom");
        }
        if (automaticPending)
        {
            return new WorkActivity(WorkPhase.ToolCall, Detail: "Collecting automatic snapshots");
        }
        if (workerPending)
        {
            ToolCallPart? call = active?.Call;
            if (call is null && nextCall > 0 && nextCall <= calls.Length)
            {
                call = calls[nextCall - 1];
            }
            if (call is not null)
            {
                return ToolActivity(call);
            }
        }
        // Foreground work takes presentation priority. Controls can overlap;
        // one retained control supplies a representative phase, not a
        // claim that the Agent has only one running Job.
        if (controlReplies.Count != 0)
        {
            return ToolActivity(controlReplies.First().Value);
        }
        if (startupReplies.Count != 0)
        {
            return new WorkActivity(WorkPhase.ToolCall, Detail: "Initializing skills");
        }
        return new WorkActivity(WorkPhase.Idle);
    }

    private static partial WorkActivity ToolActivity(ToolCallPart call)
    {
        // Describe the operation, not its arguments. Supervision must not copy
        // a command, a private question or a file's content into a status label.
        string detail = call.Name switch
        {
            "spawn_sub_agent" => "Observing delegated work",
            "join_agents" => "Waiting for child submissions",
            "shutdown_agent" => "Waiting for child quiescence",
            "ask_agent" => "Waiting for a child reply",
            "wait_for_agent" => "Waiting for child output",
            "wait_for_agent_change" => "Waiting for child progress",
            "send_message_to_parent" => "Waiting for a reply",
            "wait_for_job" => "Observing a Job",
            _ => "Executing a tool"
        };
        WorkPhase phase = call.Name is "spawn_sub_agent" or "join_agents"
            or "send_message_to_parent" or "wait_for_job" or "shutdown_agent" or "ask_agent"
            or "wait_for_agent" or "wait_for_agent_change"
            ? WorkPhase.Waiting : WorkPhase.ToolCall;
        return new WorkActivity(phase, call.Name, detail);
    }

    private partial void PublishWorkActivity()
    {
        WorkActivity current = ReadWorkActivity();
        if (current == lastEmittedActivity)
        {
            return;
        }
        lock (ingress)
        {
            // Closed status queries cannot enter the actor. They may read this
            // last published projection under the same admission lock.
            lastEmittedActivity = current;
        }
        Publish(new Observation(ObservationKind.ActivityChanged, Activity: current));
    }
}
