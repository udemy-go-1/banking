import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import TextField from '@mui/material/TextField';
import Typography from "@mui/material/Typography";
import AccountCircleIcon from '@mui/icons-material/AccountCircle';

import DefaultLayout from "../../../components/DefaultLayout";
import { getReadableDate } from "../../../src/formatUtils";
import serverSideProps from "../../../src/serverSideProps";

export async function getServerSideProps(context) {
    return await serverSideProps(context);
}

const textFieldProps = {
    input: {
        readOnly: true,
        style: {
            fontWeight: '600',
        },
    },
    label: {
        style: {
            fontWeight: '600',
        },
    },
};

export default function CustomerProfilePage(props) {
    return (
        <DefaultLayout
            homepage={props.homepage}
            authServerAddress={props.authServerAddress}
            tabTitle={"Home"}
        >
            <Box className="profile__box">
                <Paper className="profile__paper">
                    <Typography variant="h4" align="center" fontWeight="600">
                        Profile
                        <br/>
                        <AccountCircleIcon className="profile__icon"/>
                    </Typography>

                    <TextField
                        id="profile-name"
                        label="NAME"
                        autoComplete="off"
                        defaultValue={props.responseData["full_name"]}
                        InputProps={textFieldProps.input}
                        InputLabelProps={textFieldProps.label}
                        variant="standard"
                    />
                    <TextField
                        id="profile-dob"
                        label="DATE OF BIRTH"
                        autoComplete="off"
                        defaultValue={getReadableDate(props.responseData["date_of_birth"])}
                        InputProps={textFieldProps.input}
                        InputLabelProps={textFieldProps.label}
                        variant="standard"
                    />
                    <TextField
                        id="profile-email"
                        label="EMAIL"
                        autoComplete="off"
                        defaultValue={props.responseData["email"]}
                        InputProps={textFieldProps.input}
                        InputLabelProps={textFieldProps.label}
                        variant="standard"
                    />
                    <TextField
                        id="profile-country"
                        label="COUNTRY"
                        autoComplete="off"
                        defaultValue={props.responseData["country"]}
                        InputProps={textFieldProps.input}
                        InputLabelProps={textFieldProps.label}
                        variant="standard"
                    />
                    <TextField
                        id="profile-zip"
                        label="ZIP CODE"
                        autoComplete="off"
                        defaultValue={props.responseData["zipcode"]}
                        InputProps={textFieldProps.input}
                        InputLabelProps={textFieldProps.label}
                        variant="standard"
                    />
                </Paper>
            </Box>
        </DefaultLayout>
    );
}
