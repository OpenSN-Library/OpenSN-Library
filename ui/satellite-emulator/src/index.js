import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './Page/MainPage.js';
import reportWebVitals from './reportWebVitals';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import {HelpPage} from './Page/HelpPage';
import {AboutPage} from './Page/AboutPage';
import { NamespaceDetilPage } from './Page/NamespaceDetailPage.js';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <BrowserRouter>
    <Routes>
      <Route path="/help" element={<HelpPage />} />
      <Route path="/about" element={<AboutPage />} />
      <Route path="/namespace/:name/detail" element={<NamespaceDetilPage />} />
      <Route path="/node/:index" element={<App />} />
      <Route path="/" element={<App />} />
    </Routes>
  </BrowserRouter>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
