
function FormInput({ inputObj }) {



    return (
        <input className='p-2 pl-5 rounded-sm border-2 border-border focus:border-ring bg-input-background w-full' key={inputObj.id} type={inputObj.type} placeholder={inputObj.placeholder} />
    )
}

export default FormInput;