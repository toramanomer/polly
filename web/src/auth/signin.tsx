import { useMutation } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { Link, useNavigate } from 'react-router'
import {
	Box,
	Button,
	FormControl,
	FormHelperText,
	FormLabel,
	Input,
	Sheet,
	Stack,
	Typography
} from '@mui/joy'

import Email from '@mui/icons-material/Email'
import Lock from '@mui/icons-material/Lock'
import Login from '@mui/icons-material/Login'

interface SigninError {
	type: string
	title?: string
	errors?: Record<string, string[]>
}

export const Signin = () => {
	const navigate = useNavigate()

	const mutation = useMutation({
		mutationFn: async (data: any) => {
			const response = await fetch('/api/auth/signin', {
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
		onSuccess: data => {
			navigate('/home', { replace: true })
		},
		onError: (error: SigninError) => {
			if (error.type === 'invalid_credentials') {
				form.setError('root', {
					type: 'manual',
					message: 'Invalid email or password'
				})
			} else if (error.type === 'validation_error' && error.errors) {
				Object.entries(error.errors).forEach(([field, messages]) => {
					form.setError(field as any, {
						type: 'manual',
						message: messages[0]
					})
				})
			} else {
				form.setError('root', {
					type: 'manual',
					message: error.title || 'An error occurred during signin'
				})
			}
		}
	})

	const form = useForm({
		defaultValues: {
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
					<Login sx={{ fontSize: 40, color: 'primary.500', mb: 1 }} />
					<Typography level='h4' component='h1'>
						Sign In
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
							error={Boolean(form.formState.errors.email)}
						>
							<FormLabel>Email</FormLabel>
							<Input
								{...form.register('email')}
								startDecorator={<Email />}
							/>
							<FormHelperText>
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
								type='password'
							/>
							<FormHelperText>
								{form.formState.errors.password?.message}
							</FormHelperText>
						</FormControl>
						<Button
							type='submit'
							size='lg'
							loading={mutation.isPending}
						>
							Signin
						</Button>
						<Typography level='body-sm' textAlign='center'>
							Don't have an account yet?{' '}
							<Link
								to='/signup'
								style={{ textDecoration: 'none' }}
							>
								Create an account
							</Link>
						</Typography>
					</Stack>
				</form>
			</Sheet>
		</Box>
	)
}
