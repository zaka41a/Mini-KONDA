import React from "react";

export default function Header({ title }) {
  return (
    <header className="header">
      <h1>{title}</h1>
      <p>Upload a dataset and let AI explain your data semantics</p>
    </header>
  );
}
