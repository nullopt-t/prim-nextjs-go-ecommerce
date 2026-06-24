
function Button({ type }) {

    const buttonTxt = type === 'login' ? 'login'
        : type === 'register' ? 'create an account' :
            type === 'forget' ? 'continue' :
                type === 'reset' ? 'Verify OTP Code' : '';

    return (
        <input className='bg-primary text-primary-foreground rounded-sm pt-2.5 pb-2.5 font-medium capitalize hover:bg-accent hover:text-accent-foreground' type='submit' value={buttonTxt} />
    )
}

export default Button;