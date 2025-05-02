import React from 'react';

const RemediationTab = ({ reportData }) => {
    // Function to process remediation items and pair them with recommended actions
    const processRemediationItems = (items, type) => {
        if (!items || items.length === 0) return [];

        return items.map(item => {
            // Check if the item has a colon which typically separates name from observation
            const parts = item.split(':');

            if (parts.length >= 2) {
                return {
                    type: type,
                    name: parts[0].trim(),
                    observation: parts.slice(1).join(':').trim(),
                    recommendation: getRecommendation(type, parts[0].trim())
                };
            } else {
                return {
                    type: type,
                    name: '',
                    observation: item.trim(),
                    recommendation: getRecommendation(type, '')
                };
            }
        });
    };

    // Helper to generate recommendations based on item type and name
    const getRecommendation = (type, name) => {
        if (type === 'required') {
            if (name.includes('Kubeadmin')) {
                return 'Remove the kubeadmin user after confirming other admin users are set up. Use "oc delete secret kubeadmin -n kube-system" to remove the user.';
            }
            return 'This change is required for security or stability. Follow the OpenShift documentation for proper remediation steps.';
        } else if (type === 'recommended') {
            if (name.includes('Network Policy')) {
                return 'Implement network policies to control traffic between namespaces and applications. Follow zero-trust networking principles.';
            } else if (name.includes('Cluster Version')) {
                return 'Plan to update to the latest cluster version following the standard upgrade path. Test in non-production first.';
            } else if (name.includes('LimitRange')) {
                return 'Configure LimitRange objects in each namespace to control resource consumption and ensure fair resource allocation.';
            } else if (name.includes('Monitoring')) {
                return 'Enable persistent storage for monitoring components to preserve metrics history across restarts.';
            }
            return 'Consider implementing this recommendation to align with OpenShift best practices.';
        } else {
            return 'Review this information to better understand your cluster configuration.';
        }
    };

    const requiredItems = processRemediationItems(reportData.itemsRequired, 'required');
    const recommendedItems = processRemediationItems(reportData.itemsRecommended, 'recommended');

    return (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
            <h2 className="text-2xl font-bold mb-6">Remediation Actions</h2>

            <div className="mb-6">
                <h3 className="text-xl font-semibold mb-4">Required Actions ({requiredItems.length})</h3>
                {requiredItems.length > 0 ? (
                    <div className="space-y-4">
                        {requiredItems.map((item, index) => (
                            <div key={`req-${index}`} className="border border-red-200 bg-red-50 rounded-lg p-4 transition-all duration-300 hover:shadow-md">
                                <h4 className="font-medium text-red-800">Item: {item.name || 'Critical Issue'}</h4>
                                <div className="mt-2">
                                    <h5 className="font-medium text-red-800">Observation:</h5>
                                    <p className="text-gray-800 mb-2">{item.observation}</p>
                                </div>
                                <div className="mt-2">
                                    <h5 className="font-medium text-red-800">Recommendation:</h5>
                                    <p className="text-gray-800">{item.recommendation}</p>
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <p className="text-gray-500 bg-green-50 p-4 rounded-lg border border-green-100">
                        No critical changes required. Your cluster meets the basic security and stability requirements.
                    </p>
                )}
            </div>

            <div className="mb-4">
                <h3 className="text-xl font-semibold mb-4">Recommended Actions ({recommendedItems.length})</h3>
                {recommendedItems.length > 0 ? (
                    <div className="space-y-4">
                        {recommendedItems.map((item, index) => (
                            <div key={`rec-${index}`} className="border border-yellow-200 bg-yellow-50 rounded-lg p-4 transition-all duration-300 hover:shadow-md">
                                <h4 className="font-medium text-yellow-800">Item: {item.name || 'Recommendation'}</h4>
                                <div className="mt-2">
                                    <h5 className="font-medium text-yellow-800">Observation:</h5>
                                    <p className="text-gray-800 mb-2">{item.observation}</p>
                                </div>
                                <div className="mt-2">
                                    <h5 className="font-medium text-yellow-800">Recommendation:</h5>
                                    <p className="text-gray-800">{item.recommendation}</p>
                                </div>
                            </div>
                        ))}
                    </div>
                ) : (
                    <p className="text-gray-500 bg-green-50 p-4 rounded-lg border border-green-100">
                        No recommendations needed. Your cluster follows OpenShift best practices.
                    </p>
                )}
            </div>
        </div>
    );
};

export default RemediationTab;