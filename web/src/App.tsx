import { createBrowserRouter, Navigate, RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { CssVarsProvider, CssBaseline } from '@mui/joy'

import { Signin } from './auth/signin'
import { Signup } from './auth/signup'
import { Home } from './home/home'
import { Vote } from './poll/vote'

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
		element: <Home />
	},
	{
		path: '/polls/:pollId',
		element: <Vote />
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
			<CssVarsProvider defaultColorScheme='dark' defaultMode='dark'>
				<CssBaseline />
				<RouterProvider router={router} />
			</CssVarsProvider>
		</QueryClientProvider>
	)
}
