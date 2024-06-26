import Link from "next/link";
import { useRouter } from "next/router";
import { useContext, useState } from "react";
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Container from '@mui/material/Container';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import Grid from '@mui/material/Grid';
import InputAdornment from '@mui/material/InputAdornment';
import RadioGroup from '@mui/material/RadioGroup';
import TextField from '@mui/material/TextField';
import ArrowBackIosIcon from "@mui/icons-material/ArrowBackIos";

import { DataToDisplayContext } from "../../../../_app";
import ConfirmationDialog from "../../../../../components/Dialog";
import DefaultLayout from "../../../../../components/DefaultLayout";
import SnackbarAlert from "../../../../../components/SnackbarAlert";
import authServerSideProps from "../../../../../src/authServerSideProps";
import handleFetchResource from "../../../../../src/handleFetchResource";
import { validateFloat } from "../../../../../src/validationUtils";

export async function getServerSideProps(context) {
    return await authServerSideProps(context);
}

export default function TransactionPage(props) {
    const router = useRouter();

    const transactionTypeWithdrawal = "withdrawal"
    const transactionTypeDeposit = "deposit"

    const { setDataToDisplay } = useContext(DataToDisplayContext);
    const [isLoading, setIsLoading] = useState(false);
    const [selectedType, setSelectedType] = useState('');
    const [isTypeInvalid, setIsTypeInvalid] = useState(false);
    const [inputAmount, setInputAmount] = useState(0.00);
    const [isAmountInvalid, setIsAmountInvalid] = useState(false);
    const [errorMsg, setErrorMsg] = useState('');
    const [openErrorAlert, setOpenErrorAlert] = useState(false);
    const [openConfirmation, setOpenConfirmation] = useState(false);

    const accessToken = props.accessToken;
    const currentPath = props.currentPath;
    const requestURL = props.requestURL;
    const homepage = props.homepage;

    function handleSelect(e) {
        if (e.target.value !== transactionTypeWithdrawal && e.target.value !== transactionTypeDeposit) {
            setErrorMsg("Transaction type should be withdrawal or saving.");
            setOpenErrorAlert(true);
            return;
        }
        setIsTypeInvalid(false);
        setSelectedType(e.target.value);
    }

    function checkInputAmount(rawAmt) {
        const [isValid, amt] = validateFloat(rawAmt, 0.00, 10000.00);
        if (!isValid) {
            setInputAmount(0.00);
            setIsAmountInvalid(true);
            return;
        }
        setIsAmountInvalid(false);
        setOpenErrorAlert(false); //clear any previous error
        setInputAmount(amt);
    }

    function handleGetConfirmation(event) {
        event.preventDefault();
        setErrorMsg('');
        setDataToDisplay({
            isLoggingOut: false,
            pageData: [],
        }); //clear any previous data

        if (selectedType === '') {
            setIsTypeInvalid(true);
            return;
        }

        const isValid = event.target.checkValidity();
        if (!isValid || isAmountInvalid) {
            setErrorMsg("Please check that all fields are correctly filled.");
            setOpenErrorAlert(true);
        } else {
            setOpenConfirmation(true);
        }
    }

    async function handleMakeTransaction() {
        setIsLoading(true);
        setOpenConfirmation(false);

        const request = {
            method: "POST",
            headers: { "Authorization": "Bearer " + accessToken, "Content-Type": "application/json" },
            body: JSON.stringify({ "transaction_type": selectedType, "amount": inputAmount }),
        };

        const finalProps = await handleFetchResource(currentPath, requestURL, request);
        const responseData = finalProps?.props?.responseData ? [finalProps.props.responseData] : [];

        if (responseData.length === 0) {
            const possibleRedirect = finalProps?.redirect?.destination || '';
            const possibleErrInfo = finalProps?.redirect || '';
            if (possibleRedirect !== '') {
                setErrorMsg(possibleErrInfo.errorMessage);
                setOpenErrorAlert(true);
                setTimeout(() => router.replace(possibleRedirect), 10000);
            } else if (possibleErrInfo !== '' && (possibleErrInfo.statusCode === 400 || possibleErrInfo.statusCode === 422)) {
                setErrorMsg(possibleErrInfo.errorMessage);
                setOpenErrorAlert(true);
            } else {
                console.log("No response after sending transaction request");
                setErrorMsg("Something went wrong on our end, please try again later.");
                setOpenErrorAlert(true);
            }
            setIsLoading(false);
            return;
        }

        responseData.push(selectedType);
        setDataToDisplay({
            isLoggingOut: false,
            pageData: responseData,
        });

        return router.replace(`${currentPath}/success`);
    }

    return (
        <DefaultLayout
            homepage={homepage}
            authServerAddress={props.authServerAddress}
            tabTitle={"New Transaction"}
            headerTitle={"What would you like to do today?"}
        >
            <Container className="container" component="main" maxWidth="sm">
                <Box
                    component="form"
                    name="transaction-form"
                    autoComplete="off"
                    onSubmit={handleGetConfirmation}
                    height="300px"
                    align="center"
                >
                    <Grid container spacing={5}>
                        <Grid item xs={12}>
                            <FormControl required>
                                <RadioGroup id="select-transaction-type">
                                    <Grid item xs={6}>
                                        <FormControlLabel
                                            value="deposit"
                                            control={
                                                <Button
                                                    className={isTypeInvalid ? "button--type-option error" : "button--type-option"}
                                                    variant={selectedType === transactionTypeDeposit ? "contained" : "outlined"}
                                                    size="large"
                                                    onClick={handleSelect}
                                                >
                                                    Make a deposit
                                                </Button>
                                            }
                                            label=""
                                        />
                                    </Grid>
                                    <Grid item xs={6}>
                                        <FormControlLabel
                                            value="withdrawal"
                                            control={
                                                <Button
                                                    className={isTypeInvalid ? "button--type-option error" : "button--type-option"}
                                                    variant={selectedType === transactionTypeWithdrawal ? "contained" : "outlined"}
                                                    size="large"
                                                    onClick={handleSelect}
                                                >
                                                    Make a withdrawal
                                                </Button>
                                            }
                                            label=""
                                        />
                                    </Grid>
                                </RadioGroup>
                            </FormControl>
                            { isTypeInvalid &&
                                <FormHelperText className="text--align-center" error>
                                    Please select an option.
                                </FormHelperText>
                            }
                        </Grid>
                        <Grid item xs={12}>
                            <TextField
                                required
                                id="input-transaction-amount"
                                label="Amount"
                                autoComplete="off"
                                variant="standard"
                                size="large"
                                InputProps={{
                                    startAdornment: (
                                        <InputAdornment position="start">
                                            $
                                        </InputAdornment>
                                    ),
                                }}
                                inputProps={{ maxLength: 8 }}
                                error={isAmountInvalid}
                                helperText={isAmountInvalid ? "Please enter a valid amount." : "Transaction limit: $10,000"}
                                onChange={e => checkInputAmount(e.target.value)}
                            />
                        </Grid>
                        <Grid className="button--location-grid left" item xs={6}>
                            <Link href={`/customers/${router.query.id}`}>
                                <Button type="button" className="button--capitalization-off" size="small" startIcon={<ArrowBackIosIcon/>}>
                                    Go back to my accounts
                                </Button>
                            </Link>
                        </Grid>
                        <Grid className="button--location-grid right" item xs={6}>
                            <Button className="button--type-submit" type="submit" variant="contained" disabled={isLoading}>
                                {isLoading ? 'Loading...' : 'Submit'}
                            </Button>
                        </Grid>
                    </Grid>
                </Box>
            </Container>

            <ConfirmationDialog
                open={openConfirmation}
                handleNo={() => setOpenConfirmation(false)}
                handleYes={handleMakeTransaction}
                title={`Make a ${selectedType} of $${inputAmount} on Account No. ${router.query.acc_id}?`}
            />

            <SnackbarAlert
                openSnackbarAlert={openErrorAlert}
                handleClose={() => setOpenErrorAlert(false)}
                isError={true}
                title={"Transaction failed"}
                msg={errorMsg}
            />
        </DefaultLayout>
    );
}
