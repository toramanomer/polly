import { createBrowserRouter, RouterProvider } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

import { Signin } from './auth/signin'
import { Signup } from './auth/signup'

const router = createBrowserRouter([
	{
		path: '/signup',
		element: <Signup />
	},
	{
		path: '/signin',
		element: <Signin />
	}
])

const queryClient = new QueryClient()

export const App = () => (
	<QueryClientProvider client={queryClient}>
		<RouterProvider router={router} />
	</QueryClientProvider>
)
