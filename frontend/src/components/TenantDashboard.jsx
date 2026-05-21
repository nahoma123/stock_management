import React, { useState, useEffect, useMemo } from 'react';

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


// --- Main Dashboard Component ---

const TenantDashboard = () => {
    const [tenants, setTenants] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [modal, setModal] = useState({ type: null, tenant: null }); // { type: 'log' | 'expiry', tenant: object }

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
                if (modal.type === 'log' && modal.tenant.id === tenant_id) {
                    setModal(prev => ({...prev, tenant: {...prev.tenant, creation_log: (prev.tenant.creation_log || '') + log}}));
                }
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
