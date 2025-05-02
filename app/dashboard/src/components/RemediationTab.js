import React from 'react';

const RemediationTab = ({ reportData }) => {
    // Function to filter remediation items and pair them with recommended actions
    const getRemediationItems = () => {
        if (!reportData) return [];

        const items = [];

        // Process required items
        if (reportData.itemsRequired && reportData.itemsRequired.length > 0) {
            reportData.itemsRequired.forEach((item, index) => {
                if (!item.startsWith('- ') && !item.startsWith('=') && !item.startsWith('"')) {
                    items.push({
                        type: 'required',
                        observation: item,
                        recommendation: reportData.itemsRequired[index + 1] &&
                        (reportData.itemsRequired[index + 1].startsWith('- ') || reportData.itemsRequired[index + 1].startsWith('='))
                            ? reportData.itemsRequired[index + 1]
                            : 'Follow OpenShift best practices for remediation.'
                    });
                }
            });
        }

        // Process recommended items
        if (reportData.itemsRecommended && reportData.itemsRecommended.length > 0) {
            reportData.itemsRecommended.forEach((item, index) => {
                if (!item.startsWith('- ') && !item.startsWith('=') && !item.startsWith('"') && item !== 'Changes Recommended') {
                    items.push({
                        type: 'recommended',
                        observation: item,
                        recommendation: reportData.itemsRecommended[index + 1] &&
                        (reportData.itemsRecommended[index + 1].startsWith('- ') || reportData.itemsRecommended[index + 1].startsWith('='))
                            ? reportData.itemsRecommended[index + 1]
                            : 'Consider implementing this recommendation for better cluster performance.'
                    });
                }
            });
        }

        return items;
    };

    const remediationItems = getRemediationItems();

    return (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
            <h2 className="text-2xl font-bold mb-6">Remediation Actions</h2>

            <div className="mb-4">
                <h3 className="text-xl font-semibold mb-2">Required Actions ({reportData.itemsRequired?.filter(i => !i.startsWith('- ') && !i.startsWith('=')).length || 0})</h3>
                {reportData.itemsRequired?.length > 0 ? (
                    <div className="space-y-4">
                        {remediationItems.filter(item => item.type === 'required').map((item, index) => (
                            <div key={`req-${index}`} className="border border-red-200 bg-red-50 rounded-lg p-4">
                                <h4 className="font-medium text-red-800">Observation:</h4>
                                <p className="text-gray-800 mb-2">{item.observation}</p>
                                <h4 className="font-medium text-red-800">Recommendation:</h4>
                                <p className="text-gray-800">{item.recommendation}</p>
                            </div>
                        ))}
                    </div>
                ) : (
                    <p className="text-gray-500">No critical changes required</p>
                )}
            </div>

            <div className="mb-4">
                <h3 className="text-xl font-semibold mb-2">Recommended Actions ({reportData.itemsRecommended?.filter(i => !i.startsWith('- ') && !i.startsWith('=')).length || 0})</h3>
                {reportData.itemsRecommended?.length > 0 ? (
                    <div className="space-y-4">
                        {remediationItems.filter(item => item.type === 'recommended').map((item, index) => (
                            <div key={`rec-${index}`} className="border border-yellow-200 bg-yellow-50 rounded-lg p-4">
                                <h4 className="font-medium text-yellow-800">Observation:</h4>
                                <p className="text-gray-800 mb-2">{item.observation}</p>
                                <h4 className="font-medium text-yellow-800">Recommendation:</h4>
                                <p className="text-gray-800">{item.recommendation}</p>
                            </div>
                        ))}
                    </div>
                ) : (
                    <p className="text-gray-500">No recommendations</p>
                )}
            </div>
        </div>
    );
};

export default RemediationTab;