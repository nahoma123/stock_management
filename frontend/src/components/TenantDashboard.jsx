import React, { useState, useEffect, useMemo, useCallback } from 'react';

// --- Sub-components ---

const CreateTenantForm = ({ onTenantCreated }) => {
    // ... (implementation remains the same)
    const [name, setName] = useState('');
    const [subdomain, setSubdomain] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setIsSubmitting(true);
        try {
            const response = await fetch('/api/tenants', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name, subdomain }),
            });
            if (!response.ok) {
                const errData = await response.json();
                throw new Error(errData.error || 'Failed to start tenant creation.');
            }
            const newTenant = await response.json();
            onTenantCreated(newTenant);
            setName('');
            setSubdomain('');
        } catch (err) {
            setError(err.message);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="create-form">
            <h2>Create New Tenant</h2>
            <form onSubmit={handleSubmit}>
                <div className="form-group">
                    <input type="text" placeholder="Shop Name" value={name} onChange={e => setName(e.target.value)} required />
                    <input type="text" placeholder="Subdomain" value={subdomain} onChange={e => setSubdomain(e.target.value)} required />
                </div>
                <button type="submit" disabled={isSubmitting}>{isSubmitting ? 'Creating...' : 'Create Tenant'}</button>
                {error && <p className="error-message">{error}</p>}
            </form>
        </div>
    );
};

const LogModal = ({ log, onClose }) => (
    <div className="modal-backdrop" onClick={onClose}>
        <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h2>Creation Log</h2>
            <pre className="log-box">{log || "No logs yet..."}</pre>
            <button onClick={onClose} className="modal-close-btn">Close</button>
        </div>
    </div>
);

const ExpiryModal = ({ tenant, onClose, onExpirySet }) => {
    const [expiryDate, setExpiryDate] = useState(tenant.license_expiry_date ? tenant.license_expiry_date.split('T')[0] : '');

    const handleSubmit = async () => {
        await fetch(`/api/tenants/${tenant.id}/expiry`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ expiry_date: expiryDate }),
        });
        onExpirySet();
        onClose();
    };

    return (
        <div className="modal-backdrop" onClick={onClose}>
            <div className="modal-content" onClick={e => e.stopPropagation()}>
                <h2>Set Expiry for {tenant.name}</h2>
                <input type="date" value={expiryDate} onChange={e => setExpiryDate(e.target.value)} />
                <div className="modal-actions">
                    <button onClick={handleSubmit}>Set Date</button>
                    <button onClick={onClose}>Cancel</button>
                </div>
            </div>
        </div>
    );
};

const MonitoringModal = ({ tenant, onClose }) => {
    const [data, setData] = useState(null);
    const [error, setError] = useState('');

    useEffect(() => {
        fetch(`/api/tenants/${tenant.id}/monitoring`)
            .then(async response => {
                const body = await response.json();
                if (!response.ok) throw new Error(body.error || 'Monitoring is unavailable.');
                setData(body);
            })
            .catch(err => setError(err.message));
    }, [tenant.id]);

    return (
        <div className="modal-backdrop" onClick={onClose}>
            <div className="modal-content" onClick={event => event.stopPropagation()}>
                <h2>{tenant.name} Monitoring</h2>
                {error && <p className="error-message">{error}</p>}
                {!data && !error && <p>Checking tenant agent...</p>}
                {data && <>
                    <div className="metric-grid">
                        <div><span>Status</span><strong>{data.status}</strong></div>
                        <div><span>Odoo</span><strong>{data.odoo_version}</strong></div>
                        <div><span>Users</span><strong>{data.users}</strong></div>
                        <div><span>Products</span><strong>{data.products}</strong></div>
                        <div><span>Warehouses</span><strong>{data.warehouses}</strong></div>
                        <div><span>Contract</span><strong>{data.contract_version}</strong></div>
                    </div>
                    <div className="module-summary">
                        {['platform', 'tenant', 'core'].map(layer => (
                            <div key={layer}>
                                <h3>{layer} modules</h3>
                                <ul>{data.modules.filter(module => module.layer === layer).map(module => (
                                    <li key={module.name}>{module.label || module.name} <small>{module.version}</small></li>
                                ))}</ul>
                            </div>
                        ))}
                    </div>
                </>}
                <button onClick={onClose}>Close</button>
            </div>
        </div>
    );
};

const CustomizationModal = ({ tenant, onClose }) => {
    const [file, setFile] = useState(null);
    const [token, setToken] = useState(() => sessionStorage.getItem('customizationAdminToken') || '');
    const [report, setReport] = useState(null);
    const [error, setError] = useState('');
    const [busy, setBusy] = useState(false);
    const [deployed, setDeployed] = useState(false);
	const [releases, setReleases] = useState([]);

	const loadReleases = useCallback(() => fetch(`/api/tenants/${tenant.id}/customizations/releases`)
		.then(response => response.json())
		.then(body => setReleases(Array.isArray(body) ? body : [])), [tenant.id]);

	useEffect(() => {
		loadReleases();
	}, [loadReleases]);

    const submit = async (action) => {
        if (!file || !token) return;
        setBusy(true);
        setError('');
        sessionStorage.setItem('customizationAdminToken', token);
        const formData = new FormData();
        formData.append('package', file);
        try {
            const response = await fetch(`/api/tenants/${tenant.id}/customizations/${action}`, {
                method: 'POST',
                headers: { 'X-Customization-Admin-Token': token },
                body: formData,
            });
            const body = await response.json();
            if (!response.ok) throw new Error(body.error || `Could not ${action} package.`);
            setReport(body.report || body);
            if (action === 'deploy') {
				setDeployed(true);
				loadReleases();
			}
        } catch (err) {
            setError(err.message);
        } finally {
            setBusy(false);
        }
    };

	const changeRelease = async (release, action) => {
		if (!token || !window.confirm(`${action === 'rollback' ? 'Roll back to' : 'Activate'} ${release.module_name} v${release.version}?`)) return;
		setBusy(true);
		setError('');
		try {
			const response = await fetch(`/api/tenants/${tenant.id}/customizations/releases/${release.id}/${action}`, {
				method: 'POST',
				headers: { 'X-Customization-Admin-Token': token },
			});
			const body = await response.json();
			if (!response.ok) throw new Error(body.error || 'Release change failed.');
			await loadReleases();
		} catch (err) {
			setError(err.message);
		} finally {
			setBusy(false);
		}
	};

    return (
        <div className="modal-backdrop" onClick={onClose}>
            <div className="modal-content customization-modal" onClick={event => event.stopPropagation()}>
                <h2>Deploy customization to {tenant.name}</h2>
                <label>Odoo module package</label>
                <input type="file" accept=".zip,application/zip" onChange={event => { setFile(event.target.files[0]); setReport(null); setDeployed(false); }} />
                <label>Operator token</label>
                <input type="password" value={token} onChange={event => setToken(event.target.value)} autoComplete="off" />
                {error && <p className="error-message">{error}</p>}
                {report && <div className="validation-report">
                    <strong>{report.module}</strong>
                    <span>{report.files} files, {Math.ceil(report.size / 1024)} KB</span>
                    {report.warnings.map(warning => <p key={warning}>{warning}</p>)}
                </div>}
				{deployed && <p className="success-message">Package staged. Activate it from release history after review.</p>}
                <div className="modal-actions">
                    <button onClick={() => submit('validate')} disabled={!file || !token || busy}>Validate</button>
                    <button onClick={() => submit('deploy')} disabled={!report || deployed || busy}>Deploy</button>
                    <button onClick={onClose}>Close</button>
                </div>
				<h3>Release history</h3>
				<div className="release-list">
					{releases.length === 0 && <p>No customization releases yet.</p>}
					{releases.map(release => <div key={release.id}>
						<span><strong>{release.module_name || 'Pending package'} v{release.version}</strong> {release.state}</span>
						<div>
							{['staged', 'failed'].includes(release.state) && <button onClick={() => changeRelease(release, 'activate')} disabled={busy}>Activate</button>}
							{release.state === 'superseded' && <button onClick={() => changeRelease(release, 'rollback')} disabled={busy}>Roll back</button>}
						</div>
					</div>)}
				</div>
            </div>
        </div>
    );
};


// --- Main Dashboard Component ---

const TenantDashboard = () => {
    const [tenants, setTenants] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [modal, setModal] = useState({ type: null, tenant: null });

    useEffect(() => {
        // ... (fetchTenants and WebSocket logic remains the same)
        const fetchTenants = async () => {
            try {
                setLoading(true);
                const response = await fetch('/api/tenants');
                if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
                const data = await response.json();
                setTenants(data || []);
            } catch (e) {
                setError(e.message);
            } finally {
                setLoading(false);
            }
        };
        fetchTenants();

        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const ws = new WebSocket(`${wsProtocol}//${window.location.host}/ws`);

        ws.onmessage = (event) => {
            const message = JSON.parse(event.data);

            if (message.type === 'tenant_updated') {
                const updatedTenant = message.payload;
                setTenants(prev => {
                    const index = prev.findIndex(t => t.id === updatedTenant.id);
                    if (index > -1) {
                        const newTenants = [...prev];
                        newTenants[index] = { ...newTenants[index], ...updatedTenant };
                        return newTenants;
                    }
                    return [updatedTenant, ...prev];
                });
            } else if (message.type === 'tenant_log') {
                const { tenant_id, log } = message.payload;
                setTenants(prev => prev.map(t =>
                    t.id === tenant_id
                        ? { ...t, creation_log: (t.creation_log || '') + log }
                        : t
                ));
                setModal(prev => {
                    if (prev.type !== 'log' || prev.tenant?.id !== tenant_id) return prev;
                    return {...prev, tenant: {...prev.tenant, creation_log: (prev.tenant.creation_log || '') + log}};
                });
            }
        };

        return () => ws.close();
    }, []);

    const handleLifecycleAction = async (tenantId, action) => {
        if (!window.confirm(`Are you sure you want to ${action} this tenant?`)) return;
        await fetch(`/api/tenants/${tenantId}/${action}`, { method: 'POST' });
        // UI will update via WebSocket
    };

    const filteredTenants = useMemo(() => {
        // ... (filtering logic remains the same)
        return tenants.filter(tenant =>
            (tenant.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
             tenant.subdomain.toLowerCase().includes(searchTerm.toLowerCase())) &&
            (statusFilter === 'all' || tenant.state === statusFilter)
        );
    }, [tenants, searchTerm, statusFilter]);

    if (loading) return <div>Loading tenants...</div>;
    if (error) return <div>Error: {error}. Is the backend server running?</div>;

    const renderModal = () => {
        if (!modal.tenant) return null;
        if (modal.type === 'log') {
            return <LogModal log={modal.tenant.creation_log} onClose={() => setModal({type: null, tenant: null})} />;
        }
        if (modal.type === 'expiry') {
            return <ExpiryModal tenant={modal.tenant} onClose={() => setModal({type: null, tenant: null})} onExpirySet={() => {}} />;
        }
        if (modal.type === 'monitoring') {
            return <MonitoringModal tenant={modal.tenant} onClose={() => setModal({type: null, tenant: null})} />;
        }
        if (modal.type === 'customization') {
            return <CustomizationModal tenant={modal.tenant} onClose={() => setModal({type: null, tenant: null})} />;
        }
        return null;
    };

    return (
        <div className="dashboard">
            {renderModal()}
            <CreateTenantForm onTenantCreated={(newTenant) => {
                setTenants(prev => [newTenant, ...prev.filter(t => t.id !== newTenant.id)]);
            }} />
            <h1>Tenant Dashboard</h1>
            <div className="controls">
                {/* ... (search and filter controls remain the same) */}
                <input type="text" placeholder="Search by name or subdomain..." value={searchTerm} onChange={e => setSearchTerm(e.target.value)} className="search-input" />
                <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} className="filter-select">
                    <option value="all">All Statuses</option>
                    <option value="draft">Draft</option>
                    <option value="creating">Creating</option>
                    <option value="active">Active</option>
                    <option value="disabled">Disabled</option>
                    <option value="expired">Expired</option>
                    <option value="error">Error</option>
                </select>
            </div>
            <table className="tenants-table">
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Subdomain</th>
                        <th>Status</th>
                        <th>API Key</th>
                        <th>Expiry Date</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {filteredTenants.map(tenant => (
                        <tr key={tenant.id}>
                            <td>{tenant.name}</td>
                            <td>{tenant.subdomain}</td>
                            <td><span className={`status status-${tenant.state}`}>{tenant.state}</span></td>
                            <td className="api-key-cell">
                                {tenant.api_key ? (
                                    <span className="api-key" onClick={() => navigator.clipboard.writeText(tenant.api_key)} title="Click to copy">
                                        {tenant.api_key.substring(0, 15)}...
                                    </span>
                                ) : 'N/A'}
                            </td>
                            <td>{tenant.license_expiry_date ? tenant.license_expiry_date.split('T')[0] : 'N/A'}</td>
                            <td className="actions">
                                {tenant.state === 'creating' && <button onClick={() => setModal({type: 'log', tenant})}>View Log</button>}
                                {tenant.state === 'active' && <button className="action-btn-disable" onClick={() => handleLifecycleAction(tenant.id, 'disable')}>Disable</button>}
                                {tenant.state === 'disabled' && <button className="action-btn-enable" onClick={() => handleLifecycleAction(tenant.id, 'enable')}>Enable</button>}
                                {tenant.state === 'active' && <button onClick={() => setModal({type: 'monitoring', tenant})}>Monitor</button>}
                                <button onClick={() => setModal({type: 'customization', tenant})}>Customize</button>
                                <button onClick={() => setModal({type: 'expiry', tenant})}>Set Expiry</button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
};

export default TenantDashboard;
