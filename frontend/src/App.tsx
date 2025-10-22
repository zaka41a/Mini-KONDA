import React, { useState } from "react";
import * as Papa from "papaparse";
import FileUploader from "./components/FileUploader";
import AnnotationTable from "./components/AnnotationTable";
import Header from "./components/Header";
import "./index.css";

function App() {
  const [columns, setColumns] = useState([]);
  const [annotations, setAnnotations] = useState([]);
  const [loading, setLoading] = useState(false);

  const handleFileUpload = (file) => {
    Papa.parse(file, {
      header: true,
      complete: (result) => {
        setColumns(Object.keys(result.data[0]));
      },
    });
  };

  const generateAnnotations = async () => {
    setLoading(true);
    const res = await fetch("http://localhost:8080/annotate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ columns }),
    });
    const data = await res.json();
    setAnnotations(data);
    setLoading(false);
  };

  return (
    <div className="app-container">
      <Header title="Mini KONDA – Smart Data Annotator" />
      <div className="content">
        <FileUploader onUpload={handleFileUpload} />
        {columns.length > 0 && (
          <button
            className="generate-btn"
            onClick={generateAnnotations}
            disabled={loading}
          >
            {loading ? "Generating..." : "Generate Semantic Annotations"}
          </button>
        )}
        <AnnotationTable annotations={annotations} />
      </div>
    </div>
  );
}

export default App;
