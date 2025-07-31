import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './App.css'
import RecordView from './components/RecordView.jsx'
import SignUpPage from './components/SignUpPage.jsx'
import EnterToApp from './components/EnterToApp.jsx'
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';


//import App from './App.jsx'

createRoot(document.getElementById('root')).render(
  <StrictMode>
     <BrowserRouter>
      <EnterToApp />
    </BrowserRouter>
  </StrictMode>,
)
//npm run dev