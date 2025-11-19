import { useState } from 'react';

// Test Case: Form validation config drilled through form components
// Expected: 1 violation (validationRules drilled through Form → FormSection → InputField)
// Common pattern in form libraries and custom form implementations

interface ValidationRules {
    required: boolean;
    minLength?: number;
    maxLength?: number;
    pattern?: RegExp;
}

interface FormData {
    username: string;
    email: string;
}

function App() {
    const [formData, setFormData] = useState<FormData>({
        username: '',
        email: '',
    });

    const validationRules: ValidationRules = {
        required: true,
        minLength: 3,
        maxLength: 50,
        pattern: /^[a-zA-Z0-9_]+$/,
    };

    return (
        <div>
            <h1>User Registration</h1>
            <UserForm validationRules={validationRules} data={formData} onChange={setFormData} />
        </div>
    );
}

// UserForm doesn't use validationRules, just organizes form structure
interface UserFormProps {
    validationRules: ValidationRules;
    data: FormData;
    onChange: (data: FormData) => void;
}

function UserForm({ validationRules, data, onChange }: UserFormProps) {
    return (
        <form>
            <FormSection
                title="Account Information"
                validationRules={validationRules}
                data={data}
                onChange={onChange}
            />
            <button type="submit">Submit</button>
        </form>
    );
}

// FormSection doesn't use validationRules either, just groups fields
interface FormSectionProps {
    title: string;
    validationRules: ValidationRules;
    data: FormData;
    onChange: (data: FormData) => void;
}

function FormSection({ title, validationRules, data, onChange }: FormSectionProps) {
    return (
        <fieldset>
            <legend>{title}</legend>
            <InputField
                name="username"
                label="Username"
                value={data.username}
                validationRules={validationRules}
                onChange={(value) => onChange({ ...data, username: value })}
            />
            <InputField
                name="email"
                label="Email"
                value={data.email}
                validationRules={validationRules}
                onChange={(value) => onChange({ ...data, email: value })}
            />
        </fieldset>
    );
}

// InputField finally uses validationRules
interface InputFieldProps {
    name: string;
    label: string;
    value: string;
    validationRules: ValidationRules;
    onChange: (value: string) => void;
}

function InputField({ name, label, value, validationRules, onChange }: InputFieldProps) {
    const [error, setError] = useState<string | null>(null);

    const validate = (val: string) => {
        if (validationRules.required && !val) {
            setError('This field is required');
            return;
        }
        if (validationRules.minLength && val.length < validationRules.minLength) {
            setError(`Minimum length is ${validationRules.minLength}`);
            return;
        }
        if (validationRules.maxLength && val.length > validationRules.maxLength) {
            setError(`Maximum length is ${validationRules.maxLength}`);
            return;
        }
        if (validationRules.pattern && !validationRules.pattern.test(val)) {
            setError('Invalid format');
            return;
        }
        setError(null);
    };

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newValue = e.target.value;
        onChange(newValue);
        validate(newValue);
    };

    return (
        <div className="input-field">
            <label htmlFor={name}>{label}</label>
            <input
                id={name}
                type="text"
                value={value}
                onChange={handleChange}
                aria-invalid={error ? 'true' : 'false'}
            />
            {error && <span className="error">{error}</span>}
        </div>
    );
}

export default App;
