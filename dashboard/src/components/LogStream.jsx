import React, { useState, useEffect, useCallback, useRef } from "react";

const LogStream = ({ buildId }) => {
  const [logs, setLogs] = useState([]);
  const [connectionStatus, setConnectionStatus] = useState("disconnected");
  const [clientId] = useState(() => "client_" + crypto.randomUUID());
  const eventSourceRef = useRef(null);
  const logsEndRef = useRef(null);

  // Auto-scroll to bottom when new logs arrive
  const scrollToBottom = () => {
    logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [logs]);

  const connectSSE = useCallback(() => {
    if (!buildId) return;

    const url = `/v1/builds/${buildId}/logs/stream?clientId=${clientId}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setConnectionStatus("connected");
      console.log("SSE connection opened");
    };

    eventSource.onmessage = (event) => {
      // Handle generic messages
      console.log("Received message:", event.data);
    };

    eventSource.addEventListener("connected", (event) => {
      setConnectionStatus("connected");
      setLogs((prev) => [
        ...prev,
        {
          id: Date.now(),
          timestamp: new Date().toLocaleTimeString(),
          level: "info",
          message: event.data,
          type: "system",
        },
      ]);
    });

    eventSource.addEventListener("log", (event) => {
      try {
        const logEntry = JSON.parse(event.data);
        setLogs((prev) => [
          ...prev,
          {
            id: Date.now() + Math.random(),
            ...logEntry,
            type: "log",
          },
        ]);
      } catch (error) {
        // Fallback for non-JSON messages
        setLogs((prev) => [
          ...prev,
          {
            id: Date.now() + Math.random(),
            timestamp: new Date().toLocaleTimeString(),
            level: "info",
            message: event.data,
            type: "log",
          },
        ]);
      }
    });

    eventSource.onerror = (error) => {
      console.error("SSE error:", error);
      setConnectionStatus("error");

      // EventSource automatically attempts to reconnect
      // but we can show the user the connection status
      setTimeout(() => {
        if (eventSource.readyState === EventSource.CONNECTING) {
          setConnectionStatus("reconnecting");
        }
      }, 1000);
    };

    return eventSource;
  }, [buildId, clientId]);

  const disconnectSSE = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
      setConnectionStatus("disconnected");
    }
  }, []);

  // Connect on mount and buildId change
  useEffect(() => {
    const eventSource = connectSSE();

    return () => {
      if (eventSource) {
        eventSource.close();
      }
    };
  }, [connectSSE]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnectSSE();
    };
  }, [disconnectSSE]);

  const getStatusColor = () => {
    switch (connectionStatus) {
      case "connected":
        return "text-green-500";
      case "error":
        return "text-red-500";
      case "reconnecting":
        return "text-yellow-500";
      default:
        return "text-gray-500";
    }
  };

  const getLogLevelColor = (level) => {
    switch (level) {
      case "error":
        return "text-red-400";
      case "warn":
        return "text-yellow-400";
      case "info":
        return "text-blue-400";
      default:
        return "text-gray-400";
    }
  };

  const clearLogs = () => {
    setLogs([]);
  };

  return (
    <div className="log-stream bg-gray-900 text-white rounded-lg overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between bg-gray-800 px-4 py-2 border-b border-gray-700">
        <div className="flex items-center space-x-2">
          <h3 className="text-lg font-medium">Build Logs</h3>
          <span className="text-sm text-gray-400">#{buildId}</span>
        </div>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <div
              className={`w-2 h-2 rounded-full ${
                connectionStatus === "connected"
                  ? "bg-green-500"
                  : connectionStatus === "error"
                  ? "bg-red-500"
                  : connectionStatus === "reconnecting"
                  ? "bg-yellow-500"
                  : "bg-gray-500"
              }`}
            ></div>
            <span className={`text-sm ${getStatusColor()}`}>
              {connectionStatus}
            </span>
          </div>
          <button
            onClick={clearLogs}
            className="text-sm bg-gray-700 hover:bg-gray-600 px-3 py-1 rounded transition-colors"
          >
            Clear
          </button>
        </div>
      </div>

      {/* Logs Container */}
      <div className="h-96 overflow-y-auto p-4 font-mono text-sm">
        {logs.length === 0 ? (
          <div className="text-gray-500 text-center py-8">
            {connectionStatus === "connected"
              ? "Waiting for logs..."
              : "Connecting..."}
          </div>
        ) : (
          <div className="space-y-1">
            {logs.map((log) => (
              <div key={log.id} className="flex items-start space-x-2">
                <span className="text-gray-500 text-xs whitespace-nowrap">
                  {log.timestamp}
                </span>
                {log.level && (
                  <span
                    className={`text-xs font-bold uppercase ${getLogLevelColor(
                      log.level
                    )} whitespace-nowrap`}
                  >
                    {log.level}
                  </span>
                )}
                {log.phase && (
                  <span className="text-purple-400 text-xs whitespace-nowrap">
                    [{log.phase}]
                  </span>
                )}
                <span className="text-gray-100 break-all">{log.message}</span>
              </div>
            ))}
            <div ref={logsEndRef} />
          </div>
        )}
      </div>
    </div>
  );
};

export default LogStream;
