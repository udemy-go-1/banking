import { Fragment, createContext, useState } from "react";
import { CookiesProvider } from "react-cookie";

import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider } from '@mui/material/styles';

import BankTheme from '../styles/theme';
import '../styles/styles.css';

export const DataToDisplayContext = createContext(null);

export default function MyApp({ Component, pageProps }) {
    const [dataToDisplay, setDataToDisplay] = useState({
        isLoggingOut: false,
        pageData: [],
    });

    return (
        <CookiesProvider>
            <DataToDisplayContext.Provider value={{dataToDisplay: dataToDisplay, setDataToDisplay: setDataToDisplay}}>
                <ThemeProvider theme={BankTheme}>
                    <Fragment>
                        <CssBaseline />
                        <Component {...pageProps} />
                    </Fragment>
                </ThemeProvider>
            </DataToDisplayContext.Provider>
        </CookiesProvider>
    );
}
