import React, { useState } from "react";
import LogStream from "./components/LogStream";

function App() {
  const [buildId, setBuildId] = useState("demo-build-123");
  const [inputBuildId, setInputBuildId] = useState("demo-build-123");

  const handleBuildIdChange = () => {
    setBuildId(inputBuildId);
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">
          Dployr Build Logs
        </h1>

        {/* Build ID Input */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Connect to Build</h2>
          <div className="flex space-x-4">
            <input
              type="text"
              value={inputBuildId}
              onChange={(e) => setInputBuildId(e.target.value)}
              placeholder="Enter build ID"
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleBuildIdChange}
              className="px-4 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              Connect
            </button>
          </div>
          <p className="text-sm text-gray-600 mt-2">
            Current build: <span className="font-mono">{buildId}</span>
          </p>
        </div>

        {/* Log Stream Component */}
        <LogStream buildId={buildId} />

        {/* Instructions */}
        <div className="bg-white rounded-lg shadow p-6 mt-6">
          <h2 className="text-xl font-semibold mb-4">How it works</h2>
          <div className="space-y-2 text-gray-700">
            <p>
              • The LogStream component connects to{" "}
              <code className="bg-gray-100 px-2 py-1 rounded">
                /v1/builds/{"{buildId}"}/logs/stream
              </code>{" "}
              via Server-Sent Events
            </p>
            <p>
              • Each client generates a unique clientId for connection
              management
            </p>
            <p>• The connection is authenticated using your current session</p>
            <p>
              • Logs are streamed in real-time as they are generated during the
              build process
            </p>
            <p>• The connection automatically handles reconnects if it drops</p>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
