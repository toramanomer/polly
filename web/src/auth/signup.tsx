import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router'
import { useForm } from 'react-hook-form'
import {
	Box,
	Button,
	FormControl,
	FormHelperText,
	FormLabel,
	IconButton,
	Input,
	Sheet,
	Stack,
	Typography
} from '@mui/joy'

import Email from '@mui/icons-material/Email'
import Error from '@mui/icons-material/Error'
import Lock from '@mui/icons-material/Lock'
import Person from '@mui/icons-material/Person'
import PersonAdd from '@mui/icons-material/PersonAdd'
import Visibility from '@mui/icons-material/Visibility'
import VisibilityOff from '@mui/icons-material/VisibilityOff'

interface SignupError {
	type: string
	errors?: Record<string, string[]>
	title?: string
}

export const Signup = () => {
	const [passwordVisible, setPasswordVisible] = useState(false)
	const navigate = useNavigate()

	const mutation = useMutation({
		mutationFn: async (data: any) => {
			const response = await fetch('/api/auth/signup', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(data)
			})

			const responseData = await response.json()

			if (!response.ok) {
				throw responseData
			}

			return responseData
		},
		onSuccess: () => {
			navigate('/signin')
		},
		onError: (error: SignupError) => {
			if (error.type === 'email_already_exists') {
				form.setError('email', {
					type: 'manual',
					message: 'Email already exists'
				})
			} else if (error.type === 'username_already_exists') {
				form.setError('username', {
					type: 'manual',
					message: 'Username already exists'
				})
			} else if (error.type === 'validation_error' && error.errors) {
				Object.entries(error.errors).forEach(([field, messages]) => {
					form.setError(field as any, {
						type: 'manual',
						message: messages[0]
					})
				})
			} else {
				// Handle other errors
				form.setError('root', {
					type: 'manual',
					message: error.title || 'An error occurred during signup'
				})
			}
		}
	})

	const form = useForm({
		defaultValues: {
			username: '',
			email: '',
			password: ''
		}
	})

	const onSubmit = (data: any) => {
		mutation.mutate(data)
	}

	return (
		<Box
			sx={{
				display: 'flex',
				justifyContent: 'center',
				alignItems: 'center',
				height: '100vh'
			}}
		>
			<Sheet
				sx={{
					width: '500px',
					maxWidth: '90vw',
					p: 5,
					borderRadius: 'md'
				}}
			>
				<Box sx={{ textAlign: 'center', mb: 3 }}>
					<PersonAdd
						sx={{ fontSize: 60, color: 'primary.500', mb: 1 }}
					/>
					<Typography level='h4' component='h1'>
						Create Account
					</Typography>
				</Box>

				{form.formState.errors.root && (
					<Typography color='danger' sx={{ mb: 2 }}>
						{form.formState.errors.root.message}
					</Typography>
				)}
				<form onSubmit={form.handleSubmit(onSubmit)}>
					<Stack gap={3}>
						<FormControl
							size='lg'
							error={Boolean(form.formState.errors.username)}
						>
							<FormLabel>Username</FormLabel>
							<Input
								{...form.register('username')}
								startDecorator={<Person />}
							/>
							<FormHelperText
								component={Typography}
								level='title-lg'
								startDecorator={
									form.formState.errors.username ? (
										<Error />
									) : null
								}
							>
								{form.formState.errors.username?.message}
							</FormHelperText>
						</FormControl>
						<FormControl
							size='lg'
							error={Boolean(form.formState.errors.email)}
						>
							<FormLabel>Email</FormLabel>
							<Input
								{...form.register('email')}
								startDecorator={<Email />}
							/>
							<FormHelperText
								component={Typography}
								level='title-lg'
								startDecorator={
									form.formState.errors.email ? (
										<Error />
									) : null
								}
							>
								{form.formState.errors.email?.message}
							</FormHelperText>
						</FormControl>
						<FormControl
							size='lg'
							error={Boolean(form.formState.errors.password)}
						>
							<FormLabel>Password</FormLabel>
							<Input
								{...form.register('password')}
								startDecorator={<Lock />}
								endDecorator={
									<IconButton
										onClick={() =>
											setPasswordVisible(!passwordVisible)
										}
									>
										{passwordVisible ? (
											<Visibility />
										) : (
											<VisibilityOff />
										)}
									</IconButton>
								}
								type={passwordVisible ? 'text' : 'password'}
							/>
							<FormHelperText
								component={Typography}
								level='title-lg'
								startDecorator={
									form.formState.errors.password ? (
										<Error />
									) : null
								}
							>
								{form.formState.errors.password?.message}
							</FormHelperText>
						</FormControl>
						<Button
							type='submit'
							size='lg'
							loading={mutation.isPending}
						>
							Signup
						</Button>
						<Button
							onClick={() => navigate('/signin')}
							component='a'
							href='/signin'
							variant='plain'
							color='neutral'
						>
							Already have an account? Sign in
						</Button>
					</Stack>
				</form>
			</Sheet>
		</Box>
	)
}
