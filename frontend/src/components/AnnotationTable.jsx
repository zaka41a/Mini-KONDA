import React from "react";

export default function AnnotationTable({ annotations }) {
  if (!annotations.length) return null;
  return (
    <table className="annotation-table">
      <thead>
        <tr>
          <th>Column</th>
          <th>Semantic Description</th>
        </tr>
      </thead>
      <tbody>
        {annotations.map((a, i) => (
          <tr key={i}>
            <td>{a.column}</td>
            <td>{a.annotation}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
