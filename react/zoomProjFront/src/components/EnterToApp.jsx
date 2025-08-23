import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { useLocation } from 'react-router-dom';
import SignUpPage from './SignUpPage.jsx'
import SignInPage from './SignInPage.jsx'
import SignedInUserPage from './SignedInUserPage.jsx';
import MyBtn from './MyBtn.jsx';
import { useEffect } from 'react';




function EnterToApp(){

    const linkStyle = {
        color: "white",
        marginLeft: "10px"
    }
    return(
        <>
        <nav style={{position:'absolute',
            top: 0,
            left: 0,
            //width:"50px",
            
        }}>
            <Link to="/SignUpPage" style={linkStyle}>Sign Up</Link>

            <Link to="/SignInPage" style={linkStyle}>Sign In</Link>
        </nav>
        <Routes>
            <Route path='/' element={<EnterToApp />} />
            <Route path='/SignUpPage' element={<SignUpPage/>}/>
            <Route path='/SignInPage' element={<SignInPage/>}/>
            <Route path='/SignedInUserPage' element={<SignedInUserPage/>}/>
            <Route path='/MyBtn' element={<MyBtn/>}></Route>
        </Routes>

        </>
    )
}

export default EnterToApp 