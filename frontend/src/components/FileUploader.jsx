import React from "react";

export default function FileUploader({ onUpload }) {
  return (
    <div className="uploader">
      <label className="upload-label">
        📂 Upload CSV File
        <input
          type="file"
          accept=".csv"
          onChange={(e) => onUpload(e.target.files[0])}
          hidden
        />
      </label>
    </div>
  );
}
