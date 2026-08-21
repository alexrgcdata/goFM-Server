<<<<<<< HEAD
import { useEffect, useMemo, useState } from 'react'

const NAV = [['Overview', 'home'], ['Routes', 'route'], ['Metrics', 'chart'], ['Logs', 'logs'], ['Settings', 'settings']]
const initialFM = { name: '', adapter: 'dataapi', baseUrl: 'https://', database: '', resource: '', credential: '', username: '', password: '', find: true, get: true, create: false, update: false, delete: false }
const initialOther = { id: '', path: '/api/', method: 'GET', target: 'https://', auth: 'application' }

function joinUrl(base, path) { return `${base.trim().replace(/\/+$/, '')}/${path.trim().replace(/^\/+/, '')}` }
async function readResponse(response) { const text = await response.text(); let body = text; try { body = text ? JSON.parse(text) : null } catch {}; return { body, status: response.status, ok: response.ok } }

function Icon({ name, size = 20 }) {
  const paths = {
    home: <><path d="M3 11 12 3l9 8"/><path d="M5 10v10h14V10M9 20v-6h6v6"/></>, route: <><circle cx="6" cy="18" r="3"/><circle cx="18" cy="6" r="3"/><path d="M8 16 16 8M6 3v5a4 4 0 0 0 4 4h4"/></>, chart: <><path d="M4 20V10M10 20V4M16 20v-7M22 20H2"/></>, logs: <><path d="M6 3h12v18H6zM9 8h6M9 12h6M9 16h4"/></>, settings: <><circle cx="12" cy="12" r="3"/><path d="M19 15a2 2 0 0 0 0-6l1-2-3-3-2 1a2 2 0 0 0-6 0L7 4 4 7l1 2a2 2 0 0 0 0 6l-1 2 3 3 2-1a2 2 0 0 0 6 0l2 1 3-3z"/></>, bolt: <path d="m13 2-9 12h7l-1 8 9-12h-7z"/>, refresh: <><path d="M20 11a8 8 0 0 0-15-3M4 4v5h5M4 13a8 8 0 0 0 15 3M20 20v-5h-5"/></>, plus: <path d="M12 5v14M5 12h14"/>, check: <path d="m5 12 4 4L19 6"/>, close: <path d="m6 6 12 12M18 6 6 18"/>, database: <><ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v7c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 12v7c0 1.7 3.6 3 8 3s8-1.3 8-3v-7"/></>, lock: <><rect x="4" y="10" width="16" height="10" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/></>, eye: <><path d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6S2 12 2 12"/><circle cx="12" cy="12" r="2.5"/></>, chevron: <path d="m9 18 6-6-6-6"/>, copy: <><rect x="9" y="9" width="11" height="11" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></>
  }
  return <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{paths[name]}</svg>
}
function Button({ children, variant = 'primary', className = '', ...props }) { return <button className={`button ${variant} ${className}`} {...props}>{children}</button> }
function Empty({ icon = 'route', title, text }) { return <div className="empty-state"><span><Icon name={icon} size={24}/></span><h3>{title}</h3><p>{text}</p></div> }
function Field({ label, hint, children, full = false }) { return <label className={`field ${full ? 'full' : ''}`}><span>{label}{hint && <small>{hint}</small>}</span>{children}</label> }
function StatusPill({ good, children }) { return <span className={`status-pill ${good ? 'good' : 'bad'}`}><i/>{children}</span> }

export default function App() {
  const defaultGateway = import.meta.env.VITE_API_BASE_URL || (import.meta.env.DEV ? 'http://localhost:8080' : `${window.location.origin}/openbridge/go/api`)
  const [activeNav, setActiveNav] = useState('Overview'), [routeTab, setRouteTab] = useState('FileMaker')
  const [baseUrl, setBaseUrl] = useState(defaultGateway), [adminToken, setAdminToken] = useState(''), [showToken, setShowToken] = useState(false)
  const [telemetry, setTelemetry] = useState({ overview: null, routes: [], connections: [], logs: [], logSettings: null })
  const [fmForm, setFMForm] = useState(initialFM), [otherForm, setOtherForm] = useState(initialOther)
  const [logLimit, setLogLimit] = useState(50), [selectedLog, setSelectedLog] = useState(null), [notice, setNotice] = useState(null)
  const [loading, setLoading] = useState(false), [health, setHealth] = useState('unknown'), [draftToken, setDraftToken] = useState('')
  const apiOrigin = useMemo(() => { try { return new URL(baseUrl).origin } catch { return 'Not configured' } }, [baseUrl])
  const metrics = useMemo(() => {
    const logs = telemetry.logs, success = logs.filter(log => log.status < 400).length, failed = logs.length - success
    const average = logs.length ? Math.round(logs.reduce((sum, log) => sum + (log.duration_ms || 0), 0) / logs.length) : 0
    const buckets = [0,0,0,0,0,0,0]; logs.forEach(log => { const day = Math.floor((Date.now() - new Date(log.started_at).getTime()) / 86400000); if (day >= 0 && day < 7) buckets[6-day] += 1 })
    return { total: logs.length, success, failed, average, rate: logs.length ? Math.round(success / logs.length * 100) : 100, buckets }
  }, [telemetry.logs])

  async function adminRequest(path, options = {}) {
    if (!adminToken.trim()) throw new Error('Enter your administrator token in Settings first.')
    const response = await fetch(joinUrl(baseUrl, path), { ...options, headers: { Accept: 'application/json', Authorization: `Bearer ${adminToken.trim()}`, ...(options.body ? {'Content-Type':'application/json'} : {}), ...options.headers } })
    const result = await readResponse(response); if (!response.ok) throw new Error(result.body?.error?.message || `Request failed with status ${response.status}`); return result.body
  }
  async function checkHealth(showNotice = true) {
    setLoading(true)
    try { const response = await fetch(joinUrl(baseUrl, '/health'), { headers: { Accept: 'application/json' } }); setHealth(response.ok ? 'online' : 'offline'); if (!response.ok) throw new Error(`Health check returned ${response.status}`); if (showNotice) setNotice({type:'success', text:'The Go middleware is online and responding.'}) }
    catch (error) { setHealth('offline'); if (showNotice) setNotice({type:'error', text:error.message}) } finally { setLoading(false) }
  }
  async function refreshAdminData(silent = false) {
    if (!adminToken.trim()) { if (!silent) { setActiveNav('Settings'); setNotice({type:'info', text:'Connect this panel with your administrator token first.'}) } return }
    setLoading(true)
    try {
      const paths = ['/__gofm/overview','/__gofm/routes','/__gofm/filemaker-connections','/__gofm/logs','/__gofm/settings/logs']
      const [overview,routes,connections,logs,logSettings] = await Promise.all(paths.map(path => adminRequest(path)))
      setTelemetry({overview, routes:routes.routes||[], connections:connections.connections||[], logs:logs.logs||[], logSettings}); setLogLimit(logSettings.max_entries ?? 50); setHealth('online')
      if (!silent) setNotice({type:'success', text:'Dashboard data refreshed.'})
    } catch (error) { setNotice({type:'error', text:error.message}) } finally { setLoading(false) }
  }
  useEffect(() => { checkHealth(false) }, [])
  async function saveFileMakerConnection(event) {
    event.preventDefault(); const operations = ['find','get','create','update','delete'].filter(operation => fmForm[operation]), resource = fmForm.resource.trim()
    const connection = { name:fmForm.name.trim(), adapter:fmForm.adapter, base_url:fmForm.baseUrl.trim(), credential:(fmForm.credential||`${fmForm.name}-login`).trim(), default_database:fmForm.database.trim(), allowed_databases:[fmForm.database.trim()], allowed_operations:operations, ...(fmForm.adapter==='dataapi'?{allowed_layouts:[resource]}:{allowed_tables:[resource]}) }
    setLoading(true)
    try { await adminRequest('/__gofm/filemaker-connections',{method:'POST',body:JSON.stringify({connection,credential:{username:fmForm.username,password:fmForm.password}})}); setFMForm(initialFM); await refreshAdminData(true); setNotice({type:'success',text:'FileMaker connection added. Its credentials are encrypted in the server vault.'}) }
    catch(error){setNotice({type:'error',text:error.message})} finally{setLoading(false)}
  }
  async function saveOtherRoute(event) {
    event.preventDefault(); setLoading(true)
    try { await adminRequest('/__gofm/routes',{method:'POST',body:JSON.stringify({id:otherForm.id.trim(),path:otherForm.path.trim(),methods:[otherForm.method],auth:otherForm.auth,target:{type:'http',url:otherForm.target.trim()},hooks:{before:'fireGoHookBefore',after:'fireGoHookAfter'}})}); setOtherForm(initialOther); await refreshAdminData(true); setNotice({type:'success',text:'Route is active for this server session. Add it to config.json before restarting.'}) }
    catch(error){setNotice({type:'error',text:error.message})} finally{setLoading(false)}
  }
  async function saveLogSettings(){setLoading(true);try{await adminRequest('/__gofm/settings/logs',{method:'PUT',body:JSON.stringify({max_entries:Number(logLimit)})});await refreshAdminData(true);setNotice({type:'success',text:logLimit===0?'Transaction logging is disabled.':`The server will retain the latest ${logLimit} transactions.`})}catch(error){setNotice({type:'error',text:error.message})}finally{setLoading(false)}}
  function generateToken(){const bytes=new Uint8Array(32);crypto.getRandomValues(bytes);setDraftToken(btoa(String.fromCharCode(...bytes)));setNotice({type:'info',text:'Token draft generated. Add it to the server environment before it can authenticate requests.'})}
  function navigate(item){setActiveNav(item);if(item!=='Settings'&&adminToken.trim())refreshAdminData(true);window.scrollTo({top:0,behavior:'smooth'})}
  const connected=Boolean(adminToken.trim()&&telemetry.overview), title={Overview:'Good evening, Alex.',Routes:'Build secure routes.',Metrics:'See what is happening.',Logs:'Inspect every transaction.',Settings:'Secure your gateway.'}[activeNav]

  return <div className="app-shell">
    <header className="app-header"><button className="brand" onClick={()=>navigate('Overview')}><span className="brand-mark"><Icon name="bolt" size={18}/></span><span>goFM <small>middleware console</small></span></button><div className="header-actions"><StatusPill good={health==='online'}>{health==='online'?'Server online':health==='offline'?'Server unavailable':'Checking server'}</StatusPill><button className="avatar" onClick={()=>navigate('Settings')} aria-label="Open settings">AS</button></div></header>
    <main className="content">
      {notice&&<div className={`notice ${notice.type}`}><span>{notice.type==='success'?<Icon name="check" size={18}/>:<Icon name="lock" size={18}/>} {notice.text}</span><button onClick={()=>setNotice(null)} aria-label="Dismiss"><Icon name="close" size={16}/></button></div>}
      <section className="page-heading"><div><p className="eyebrow">{activeNav==='Routes'?'Connection workspace':'FileMaker to web gateway'}</p><h1>{title}</h1><p>One secure place for routes, credentials, activity, and troubleshooting.</p></div><Button variant="secondary" onClick={()=>refreshAdminData()} disabled={loading}><Icon name="refresh" size={17}/>{loading?'Working…':'Refresh data'}</Button></section>
      {activeNav==='Overview'&&<Overview health={health} apiOrigin={apiOrigin} telemetry={telemetry} metrics={metrics} connected={connected} onNavigate={navigate} onHealth={checkHealth}/>} 
      {activeNav==='Routes'&&<Routes routeTab={routeTab} setRouteTab={setRouteTab} fmForm={fmForm} setFMForm={setFMForm} otherForm={otherForm} setOtherForm={setOtherForm} saveFM={saveFileMakerConnection} saveOther={saveOtherRoute} telemetry={telemetry} connected={connected} loading={loading} onSettings={()=>navigate('Settings')}/>} 
      {activeNav==='Metrics'&&<Metrics metrics={metrics} logs={telemetry.logs} connected={connected} onSettings={()=>navigate('Settings')}/>} 
      {activeNav==='Logs'&&<Logs logs={telemetry.logs} settings={telemetry.logSettings} logLimit={logLimit} setLogLimit={setLogLimit} saveSettings={saveLogSettings} selectedLog={selectedLog} setSelectedLog={setSelectedLog} connected={connected} loading={loading} onSettings={()=>navigate('Settings')}/>} 
      {activeNav==='Settings'&&<Settings baseUrl={baseUrl} setBaseUrl={setBaseUrl} token={adminToken} setToken={setAdminToken} showToken={showToken} setShowToken={setShowToken} connect={()=>refreshAdminData()} connected={connected} overview={telemetry.overview} draftToken={draftToken} generateToken={generateToken}/>} 
      <footer>goFM middleware · Built by Alex Seidler · <a href="mailto:agseidler@gmail.com">agseidler@gmail.com</a></footer>
    </main>
    <nav className="bottom-nav" aria-label="Primary navigation">{NAV.map(([item,icon])=><button key={item} className={activeNav===item?'active':''} onClick={()=>navigate(item)}><span><Icon name={icon} size={21}/></span><small>{item}</small></button>)}</nav>
  </div>
}

function Overview({health,apiOrigin,telemetry,metrics,connected,onNavigate,onHealth}){return <div className="view-stack"><section className="stats-grid"><article className="stat-card featured"><div className="stat-label"><span>Gateway status</span><StatusPill good={health==='online'}>{health==='online'?'Healthy':'Needs attention'}</StatusPill></div><strong>{health==='online'?'Operational':'Not connected'}</strong><small>{apiOrigin}</small><div className="mini-bars">{[28,38,32,55,44,68,58,75,70].map((height,index)=><i key={index} style={{height:`${height}%`}}/>)}</div></article><Stat icon="route" color="purple" label="Active routes" value={(telemetry.overview?.route_count||0)+(telemetry.overview?.filemaker_connection_count||0)} detail={`${telemetry.connections.length} FileMaker · ${telemetry.routes.length} other`}/><Stat icon="check" color="green" label="Success rate" value={`${metrics.rate}%`} detail={`${metrics.success} successful transactions`}/><Stat icon="logs" color="amber" label="Saved logs" value={telemetry.overview?.log_count??0} detail={`of ${telemetry.overview?.log_capacity??50} retained`}/></section><section className="overview-grid"><article className="panel quick-start"><div className="panel-heading"><div><p className="eyebrow">Quick start</p><h2>Your gateway at a glance</h2></div><span className="soft-icon"><Icon name="bolt"/></span></div><div className="step-list"><Step n="1" title="Connect the admin panel" text="Enter the administrator token stored on your server." onClick={()=>onNavigate('Settings')}/><Step n="2" title="Add a secure route" text="Choose FileMaker Data API, OData, or another service." onClick={()=>onNavigate('Routes')}/><Step n="3" title="Test the middleware" text="Confirm that Go is responding before sending live traffic." onClick={()=>onHealth(true)}/></div></article><article className="panel recent-card"><div className="panel-heading"><div><p className="eyebrow">Recent activity</p><h2>Latest transactions</h2></div><button className="text-button" onClick={()=>onNavigate('Logs')}>View all</button></div>{connected&&telemetry.logs.length?telemetry.logs.slice(0,4).map(log=><div className="activity-row" key={log.id}><span className={`activity-dot ${log.status>=400?'failed':''}`}/><div><strong>{log.method} {log.path}</strong><small>{log.duration_ms} ms · {new Date(log.started_at).toLocaleString()}</small></div><b>{log.status}</b></div>):<Empty icon="logs" title="No activity yet" text={connected?'Transactions will appear as your routes are used.':'Connect the admin panel to load secure activity.'}/>}</article></section></div>}
function Stat({icon,color,label,value,detail}){return <article className="stat-card"><div className={`stat-icon ${color}`}><Icon name={icon}/></div><span>{label}</span><strong>{value}</strong><small>{detail}</small></article>}
function Step({n,title,text,onClick}){return <button onClick={onClick}><b>{n}</b><span><strong>{title}</strong><small>{text}</small></span><Icon name="chevron"/></button>}

function Routes({routeTab,setRouteTab,fmForm,setFMForm,otherForm,setOtherForm,saveFM,saveOther,telemetry,connected,loading,onSettings}){
  const toggle=operation=>setFMForm({...fmForm,[operation]:!fmForm[operation]})
  return <div className="view-stack"><div className="subnav"><button className={routeTab==='FileMaker'?'active':''} onClick={()=>setRouteTab('FileMaker')}><Icon name="database" size={18}/>FileMaker</button><button className={routeTab==='Other'?'active':''} onClick={()=>setRouteTab('Other')}><Icon name="route" size={18}/>Other services</button></div>{!connected&&<ConnectionRequired onSettings={onSettings}/>} {connected&&routeTab==='FileMaker'&&<div className="builder-layout"><form className="panel route-form" onSubmit={saveFM}><div className="panel-heading"><div><p className="eyebrow">New FileMaker route</p><h2>Connect a database securely</h2></div><span className="secure-badge"><Icon name="lock" size={15}/>Encrypted vault</span></div><p className="panel-intro">Credentials are sent once to your private Go service over HTTPS, encrypted at rest, and never returned to this browser.</p><div className="form-grid"><Field label="Route name" hint="Friendly internal name"><input required value={fmForm.name} onChange={e=>setFMForm({...fmForm,name:e.target.value})} placeholder="Customer CRM"/></Field><Field label="Connection method"><div className="segmented"><button type="button" className={fmForm.adapter==='dataapi'?'active':''} onClick={()=>setFMForm({...fmForm,adapter:'dataapi'})}>Data API</button><button type="button" className={fmForm.adapter==='odata'?'active':''} onClick={()=>setFMForm({...fmForm,adapter:'odata'})}>OData</button></div></Field><Field label="FileMaker Server URL" hint="HTTPS only" full><input required type="url" value={fmForm.baseUrl} onChange={e=>setFMForm({...fmForm,baseUrl:e.target.value})} placeholder="https://filemaker.example.com"/></Field><Field label="Database name"><input required value={fmForm.database} onChange={e=>setFMForm({...fmForm,database:e.target.value})} placeholder="CRM"/></Field><Field label={fmForm.adapter==='dataapi'?'Allowed layout':'Allowed table'}><input required value={fmForm.resource} onChange={e=>setFMForm({...fmForm,resource:e.target.value})} placeholder={fmForm.adapter==='dataapi'?'Customers':'customers'}/></Field><Field label="Credential label" hint="No spaces or slashes"><input value={fmForm.credential} onChange={e=>setFMForm({...fmForm,credential:e.target.value})} placeholder="customer-crm-login"/></Field><div/><Field label="FileMaker username"><input required autoComplete="username" value={fmForm.username} onChange={e=>setFMForm({...fmForm,username:e.target.value})} placeholder="API account"/></Field><Field label="FileMaker password"><input required type="password" autoComplete="new-password" value={fmForm.password} onChange={e=>setFMForm({...fmForm,password:e.target.value})} placeholder="Enter once; never displayed again"/></Field></div><fieldset className="permissions"><legend>Allowed operations</legend><p>Start read-only. Enable changes only when this route requires them.</p><div>{['find','get','create','update','delete'].map(op=><label key={op} className={fmForm[op]?'checked':''}><input type="checkbox" checked={fmForm[op]} onChange={()=>toggle(op)}/><span><Icon name="check" size={14}/></span>{op}</label>)}</div></fieldset><div className="form-footer"><span><Icon name="lock" size={16}/>Username and password use AES-256-GCM encryption.</span><Button type="submit" disabled={loading}><Icon name="plus" size={17}/>Add FileMaker route</Button></div></form><RouteList title="FileMaker connections" items={telemetry.connections} type="filemaker"/></div>}{connected&&routeTab==='Other'&&<div className="builder-layout"><form className="panel route-form" onSubmit={saveOther}><div className="panel-heading"><div><p className="eyebrow">New service route</p><h2>Route to another approved service</h2></div><span className="secure-badge"><Icon name="lock" size={15}/>Bearer protected</span></div><p className="panel-intro">Only server-approved destination hosts are accepted. Caller authorization is removed before forwarding.</p><div className="form-grid"><Field label="Route name" hint="Internal identifier"><input required value={otherForm.id} onChange={e=>setOtherForm({...otherForm,id:e.target.value})} placeholder="wordpress-orders"/></Field><Field label="Allowed request method"><select value={otherForm.method} onChange={e=>setOtherForm({...otherForm,method:e.target.value})}>{['GET','POST','PUT','PATCH','DELETE'].map(method=><option key={method}>{method}</option>)}</select></Field><Field label="Public route path" hint="What your apps call"><input required value={otherForm.path} onChange={e=>setOtherForm({...otherForm,path:e.target.value})} placeholder="/api/orders"/></Field><Field label="Who can call it"><select value={otherForm.auth} onChange={e=>setOtherForm({...otherForm,auth:e.target.value})}><option value="application">Application token</option><option value="admin">Administrator token</option></select></Field><Field label="Destination service URL" hint="Must be server allow-listed" full><input required type="url" value={otherForm.target} onChange={e=>setOtherForm({...otherForm,target:e.target.value})} placeholder="https://api.example.com/orders"/></Field></div><div className="hook-note"><Icon name="route" size={18}/><div><strong>Before and after hooks are reserved</strong><p>The route records <code>fireGoHookBefore</code> and <code>fireGoHookAfter</code> metadata. It never executes scripts or commands.</p></div></div><div className="form-footer"><span>Runtime routes must be copied to config.json to survive restart.</span><Button type="submit" disabled={loading}><Icon name="plus" size={17}/>Add service route</Button></div></form><RouteList title="Other service routes" items={telemetry.routes} type="other"/></div>}</div>
}
function RouteList({title,items,type}){return <aside className="panel route-list"><div className="panel-heading"><div><p className="eyebrow">Configured</p><h2>{title}</h2></div><span className="count-badge">{items.length}</span></div>{items.length?items.map((item,index)=><article className="route-card" key={item.id||item.name||index}><div className="route-card-top"><span className={`route-icon ${type}`}><Icon name={type==='filemaker'?'database':'route'} size={18}/></span><div><strong>{item.name||item.id}</strong><small>{type==='filemaker'?(item.adapter==='dataapi'?'FileMaker Data API':'FileMaker OData'):(item.methods||[]).join(', ')}</small></div><StatusPill good>Active</StatusPill></div><dl>{type==='filemaker'?<><div><dt>Database</dt><dd>{item.default_database||'Configured'}</dd></div><div><dt>Operations</dt><dd>{(item.allowed_operations||[]).join(', ')}</dd></div></>:<><div><dt>Public path</dt><dd>{item.path}</dd></div><div><dt>Destination</dt><dd>{item.target}</dd></div></>}<div><dt>Persistence</dt><dd>{item.persistent?'Saved':'Until restart'}</dd></div></dl></article>):<Empty icon={type==='filemaker'?'database':'route'} title="No routes yet" text="Use the labeled form to add your first secure route."/>}</aside>}

function Metrics({metrics,logs,connected,onSettings}){if(!connected)return <ConnectionRequired onSettings={onSettings}/>;const max=Math.max(...metrics.buckets,1),counts=logs.reduce((a,l)=>({...a,[l.route||l.path]:(a[l.route||l.path]||0)+1}),{}),popular=Object.entries(counts).sort((a,b)=>b[1]-a[1]).slice(0,5);return <div className="view-stack"><section className="metric-cards"><Metric label="Total requests" value={metrics.total} text="Current retained window"/><Metric label="Successful" value={metrics.success} text={`${metrics.rate}% success rate`} color="green-text"/><Metric label="Failed" value={metrics.failed} text="Requests requiring review" color="red-text"/><Metric label="Average time" value={metrics.average} suffix="ms" text="Gateway response duration"/></section><section className="metrics-grid"><article className="panel chart-panel"><div className="panel-heading"><div><p className="eyebrow">Seven-day activity</p><h2>Transaction volume</h2></div><span className="method-chip">Latest logs</span></div><div className="bar-chart">{metrics.buckets.map((value,index)=><div key={index}><span>{value||''}</span><i style={{height:`${Math.max(4,value/max*100)}%`}}/><small>{['-6d','-5d','-4d','-3d','-2d','Yesterday','Today'][index]}</small></div>)}</div></article><article className="panel route-performance"><div className="panel-heading"><div><p className="eyebrow">Traffic</p><h2>Busiest routes</h2></div></div>{popular.length?popular.map(([route,count])=><div className="performance-row" key={route}><span><strong>{route}</strong><small>{count} transaction{count===1?'':'s'}</small></span><div><i style={{width:`${Math.max(10,count/popular[0][1]*100)}%`}}/></div></div>):<Empty icon="chart" title="No metrics yet" text="Charts will populate as the middleware handles traffic."/>}</article></section></div>}
function Metric({label,value,text,color='',suffix}){return <article><span>{label}</span><strong className={color}>{value}{suffix&&<em>{suffix}</em>}</strong><small>{text}</small></article>}

function Logs({logs,settings,logLimit,setLogLimit,saveSettings,selectedLog,setSelectedLog,connected,loading,onSettings}){if(!connected)return <ConnectionRequired onSettings={onSettings}/>;return <div className="view-stack"><section className="panel log-controls"><div><span className="soft-icon"><Icon name="logs"/></span><div><h2>Transaction logging</h2><p>Keep a bounded history for debugging. Set this to zero to turn logging off.</p></div></div><div className="limit-control"><label>Keep latest <strong>{logLimit}</strong></label><input type="range" min="0" max="100" step="5" value={logLimit} onChange={e=>setLogLimit(Number(e.target.value))}/><Button onClick={saveSettings} disabled={loading}>Save setting</Button></div><div className="encryption-state"><Icon name="lock" size={17}/><span><strong>{settings?.encrypted?'Encrypted storage enabled':'Memory-only logging'}</strong><small>{settings?.encrypted?'Log files are encrypted at rest.':'Configure GOFM_LOG_KEY for production encryption.'}</small></span></div></section><section className="logs-layout"><article className="panel log-table"><div className="table-head"><span>Status</span><span>Request</span><span>Duration</span><span>Time</span></div>{logs.length?logs.map(log=><button className={`table-row ${selectedLog?.id===log.id?'selected':''}`} key={log.id} onClick={()=>setSelectedLog(log)}><span><b className={log.status>=400?'status-code failed':'status-code'}>{log.status}</b></span><span><strong>{log.method} {log.path}</strong><small>{log.id}</small></span><span>{log.duration_ms} ms</span><span>{new Date(log.started_at).toLocaleString()}</span></button>):<Empty icon="logs" title="No saved transactions" text={logLimit===0?'Logging is currently disabled.':'Requests will appear here as routes receive traffic.'}/>}</article><aside className="panel log-detail"><div className="panel-heading"><div><p className="eyebrow">Response viewer</p><h2>Transaction detail</h2></div>{selectedLog&&<button className="icon-button" onClick={()=>setSelectedLog(null)}><Icon name="close" size={16}/></button>}</div>{selectedLog?<><dl><div><dt>Request ID</dt><dd>{selectedLog.id}</dd></div><div><dt>Outcome</dt><dd><StatusPill good={selectedLog.status<400}>{selectedLog.outcome}</StatusPill></dd></div><div><dt>Route</dt><dd>{selectedLog.route||selectedLog.path}</dd></div><div><dt>Upstream</dt><dd>{selectedLog.upstream||'Internal gateway'}</dd></div></dl><div className="json-view"><div><span>Response preview</span><button onClick={()=>navigator.clipboard?.writeText(selectedLog.response_preview||'')}><Icon name="copy" size={15}/>Copy</button></div><pre>{selectedLog.response_preview||'Response preview storage is disabled. Enable capture_body_preview in config.json to retain sanitized response previews.'}</pre></div></>:<Empty icon="eye" title="Select a transaction" text="Choose a log entry to inspect safe metadata and response preview."/>}</aside></section></div>}

function Settings({baseUrl,setBaseUrl,token,setToken,showToken,setShowToken,connect,connected,overview,draftToken,generateToken}){return <div className="settings-grid"><section className="panel settings-card"><div className="panel-heading"><div><p className="eyebrow">Admin connection</p><h2>Connect this dashboard</h2></div><StatusPill good={connected}>{connected?'Connected':'Not connected'}</StatusPill></div><p className="panel-intro">This token manages routes, credentials, logs, and metrics. It is held only in this browser tab.</p><Field label="Go middleware URL" full><input type="url" value={baseUrl} onChange={e=>setBaseUrl(e.target.value)} placeholder="https://example.com/openbridge/go/api"/></Field><Field label="Administrator token" full><div className="password-input"><input type={showToken?'text':'password'} value={token} onChange={e=>setToken(e.target.value)} placeholder="Paste GOFM_ADMIN_TOKEN" autoComplete="off"/><button type="button" onClick={()=>setShowToken(!showToken)} aria-label="Show or hide token"><Icon name="eye" size={18}/></button></div></Field><Button className="wide" onClick={connect}><Icon name="lock" size={17}/>Connect securely</Button></section><section className="panel settings-card"><div className="panel-heading"><div><p className="eyebrow">Deployment tokens</p><h2>Generate a secure token</h2></div><span className="soft-icon"><Icon name="lock"/></span></div><p className="panel-intro">Generate a strong draft, then place it in your hosting environment as <code>GOFM_APP_TOKEN</code> or <code>GOFM_ADMIN_TOKEN</code>. Generating it here does not activate it.</p><Button variant="secondary" onClick={generateToken}>Generate 256-bit token</Button>{draftToken&&<div className="generated-token"><code>{draftToken}</code><button onClick={()=>navigator.clipboard?.writeText(draftToken)}><Icon name="copy" size={16}/>Copy</button></div>}<div className="token-rule"><Icon name="lock" size={17}/><span><strong>Keep two different tokens</strong><small>Application tokens call routes. Administrator tokens change configuration.</small></span></div></section><section className="panel security-summary"><div className="panel-heading"><div><p className="eyebrow">Security posture</p><h2>Server protections</h2></div></div><div className="security-grid"><SecurityItem label="Credential vault" ready={overview?.credential_vault} text="AES-256-GCM encrypted credentials"/><SecurityItem label="Encrypted logs" ready={overview?.logs_encrypted} text="Bounded encrypted request history"/><SecurityItem label="SQLite storage" ready={overview?.persistent_storage} text="Server-side transaction database"/><SecurityItem label="Fixed destinations" ready text="No caller-provided upstream URLs"/></div></section></div>}
function SecurityItem({label,ready,text}){return <div className="security-item"><span className={ready?'ready':''}><Icon name={ready?'check':'lock'} size={16}/></span><div><strong>{label}</strong><small>{text}</small></div><b>{ready?'Ready':'Setup needed'}</b></div>}
function ConnectionRequired({onSettings}){return <section className="panel connection-required"><span><Icon name="lock" size={28}/></span><h2>Connect the administrator panel</h2><p>Enter your Go administrator token once to manage this private server. FileMaker credentials use a separate encrypted vault.</p><Button onClick={onSettings}>Open secure settings</Button></section>}
=======
import { useState } from 'react'

function joinUrl(baseUrl, path) {
  const normalizedBase = baseUrl.trim().replace(/\/+$/, '')
  const normalizedPath = path.trim().replace(/^\/+/, '')
  return `${normalizedBase}/${normalizedPath}`
}

async function readResponse(response) {
  const body = await response.text()
  if (!body) return null

  try {
    return JSON.parse(body)
  } catch {
    return body
  }
}

export default function App() {
  const [baseUrl, setBaseUrl] = useState(import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080')
  const [token, setToken] = useState('')
  const [endpoint, setEndpoint] = useState('/api/example')
  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)

  async function request(path, needsAuth) {
    if (!baseUrl.trim()) {
      setResult({ error: 'Enter the API base URL first.' })
      return
    }

    setLoading(true)
    setResult(null)
    const url = joinUrl(baseUrl, path)
    const headers = { Accept: 'application/json' }
    if (needsAuth && token.trim()) headers.Authorization = `Bearer ${token.trim()}`

    try {
      const response = await fetch(url, { headers })
      setResult({
        url,
        status: response.status,
        ok: response.ok,
        body: await readResponse(response),
      })
    } catch (error) {
      setResult({ url, error: error.message })
    } finally {
      setLoading(false)
    }
  }

  return (
    <main>
      <section className="card" aria-labelledby="title">
        <p className="eyebrow">Example React client</p>
        <h1 id="title">goFM Server</h1>
        <p className="intro">Use this small client to check the server health endpoint or make a bearer-token API request.</p>
        <p className="cors-note">
          Browser origin: <code>{window.location.origin}</code>. Add this exact origin to the server's allowed CORS origins before calling an API on another origin.
        </p>

        <label>
          API base URL
          <input value={baseUrl} onChange={(event) => setBaseUrl(event.target.value)} placeholder="http://localhost:8080" inputMode="url" />
        </label>
        <label>
          Bearer token <span>(only sent with the API request)</span>
          <input value={token} onChange={(event) => setToken(event.target.value)} placeholder="development token" type="password" autoComplete="off" />
        </label>
        <label>
          API endpoint
          <input value={endpoint} onChange={(event) => setEndpoint(event.target.value)} placeholder="/api/example" />
        </label>

        <div className="actions">
          <button type="button" onClick={() => request('/health', false)} disabled={loading}>Check health</button>
          <button type="button" className="primary" onClick={() => request(endpoint, true)} disabled={loading}>Call API endpoint</button>
        </div>

        <section className="result" aria-live="polite" aria-label="Response">
          <h2>Response</h2>
          <pre>{result ? JSON.stringify(result, null, 2) : 'No request sent yet.'}</pre>
        </section>
      </section>
    </main>
  )
}
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
