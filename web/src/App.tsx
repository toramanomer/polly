import {
	createBrowserRouter,
	Navigate,
	RouterProvider,
	useNavigate
} from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { Signin } from './auth/signin'
import { Signup } from './auth/signup'
import { AuthProvider, useAuth } from './auth/authContext'

const LogoutButton = () => {
	const { setToken } = useAuth()
	const navigate = useNavigate()

	const handleLogout = () => {
		// Clear token from memory
		setToken(null)

		// Call logout endpoint to clear httpOnly cookie
		fetch('/api/auth/logout', {
			method: 'POST',
			credentials: 'include'
		})
			.then(() => {
				navigate('/signin', { replace: true })
			})
			.catch(() => {
				navigate('/signin', { replace: true })
			})
	}

	return <button onClick={handleLogout}>Logout</button>
}

const PrivateRoute = ({ children }: { children: React.ReactNode }) => {
	const { isAuthenticated } = useAuth()

	if (!isAuthenticated) {
		return <Navigate to='/signin' replace />
	}

	return <>{children}</>
}

const AppRoutes = () => {
	return (
		<AuthProvider>
			<RouterProvider router={router} />
		</AuthProvider>
	)
}

const router = createBrowserRouter([
	{
		path: '/signup',
		element: <Signup />
	},
	{
		path: '/signin',
		element: <Signin />
	},
	{
		path: '/home',
		element: (
			<PrivateRoute>
				<div>
					Home Page
					<LogoutButton />
				</div>
			</PrivateRoute>
		)
	},
	{
		path: '/',
		element: <Navigate to='/home' replace />
	}
])

const queryClient = new QueryClient()

export const App = () => {
	return (
		<QueryClientProvider client={queryClient}>
			<AppRoutes />
		</QueryClientProvider>
	)
}
