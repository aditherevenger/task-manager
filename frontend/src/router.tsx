import { createBrowserRouter } from 'react-router-dom'
import Layout from './components/Layout'
import TasksPage from './pages/TasksPage'
import StatsPage from './pages/StatsPage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, element: <TasksPage /> },
      { path: 'stats', element: <StatsPage /> },
    ],
  },
])
