async function post(path, body) {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  })
  return res.json()
}

export const listJobs = (params) => post('/admin/jobs/list', params)
export const disableJob = (jobID) => post('/admin/jobs/disable', { job_id: jobID })
export const enableJob = (jobID) => post('/admin/jobs/enable', { job_id: jobID })
export const deleteJob = (jobID) => post('/admin/jobs/delete', { job_id: jobID })
