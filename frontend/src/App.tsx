import { Routes, Route } from 'react-router-dom'
import { Layout } from './components/Layout'
import { ProjectsPage } from './pages/ProjectsPage'
import { ProjectPage } from './pages/ProjectPage'
import { BuildPage } from './pages/BuildPage'
import { ComparePage } from './pages/ComparePage'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<ProjectsPage />} />
        <Route path="/projects/:slug" element={<ProjectPage />} />
        <Route path="/builds/:buildId" element={<BuildPage />} />
        <Route path="/compare" element={<ComparePage />} />
      </Routes>
    </Layout>
  )
}

export default App
