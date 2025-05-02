import React, { useState } from 'react';
import { getScoreRating } from '../utils/scoreUtils';

const ExecutiveSummaryTab = ({ reportData }) => {
    const [customerName, setCustomerName] = useState('');
    const [generating, setGenerating] = useState(false);

    // Function to download executive summary as ADOC
    const generateExecutiveSummary = () => {
        if (!reportData) return;

        setGenerating(true);

        try {
            const customerText = customerName ? customerName : 'Your Company';

            const summaryContent = `= OpenShift Health Check Executive Summary
${customerText}
:toc: macro
:toc-title: Table of Contents
:toclevels: 3
:sectnums:
:sectlinks:
:icons: font

toc::[]

== Overview

Red Hat Consulting conducted a health check for ${customerText}'s OpenShift cluster. The overall cluster health is *${reportData.overallScore}%*, which is considered ${getScoreRating(reportData.overallScore)}.

== Category Health Assessment

=== Infrastructure Setup: ${reportData.scoreInfra}%
${reportData.infraDescription}

=== Policy Governance: ${reportData.scoreGovernance}%
${reportData.governanceDescription}

=== Compliance Benchmarking: ${reportData.scoreCompliance}%
${reportData.complianceDescription}

=== Central Monitoring and Logging: ${reportData.scoreMonitoring}%
${reportData.monitoringDescription}

=== Build/Deploy Security: ${reportData.scoreBuildSecurity}%
${reportData.buildSecurityDescription}

== Priority-Based Actions

=== Changes Required
${reportData.itemsRequired && reportData.itemsRequired.length > 0
                ? reportData.itemsRequired.map(item => `* ${item}`).join('\n')
                : 'No critical changes required.'}

=== Changes Recommended
${reportData.itemsRecommended && reportData.itemsRecommended.length > 0
                ? reportData.itemsRecommended.map(item => `* ${item}`).join('\n')
                : 'No recommendations.'}

=== Advisory Actions
${reportData.itemsAdvisory && reportData.itemsAdvisory.length > 0
                ? reportData.itemsAdvisory.map(item => `* ${item}`).join('\n')
                : 'No advisory actions.'}
`;

            // Create a blob and trigger download
            const blob = new Blob([summaryContent], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `exec-summary-${new Date().toISOString().split('T')[0]}.adoc`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

        } catch (error) {
            console.error('Error generating summary:', error);
            alert('Failed to generate executive summary');
        } finally {
            setGenerating(false);
        }
    };

    return (
        <div className="bg-white rounded-lg shadow p-6 mb-6">
            <h2 className="text-2xl font-bold mb-6">Executive Summary Generator</h2>

            <div className="mb-6">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                    Customer Name
                </label>
                <input
                    type="text"
                    value={customerName}
                    onChange={(e) => setCustomerName(e.target.value)}
                    placeholder="Enter customer name"
                    className="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
            </div>

            <div className="mb-6">
                <h3 className="text-lg font-semibold mb-4">Summary Preview</h3>
                <div className="bg-gray-50 p-4 rounded-md">
                    <h4 className="font-bold">OpenShift Health Check Executive Summary</h4>
                    <p className="text-gray-700">For: {customerName || 'Your Company'}</p>
                    <p className="text-gray-700">Overall Health Score: {reportData.overallScore}% ({getScoreRating(reportData.overallScore)})</p>

                    <div className="mt-4">
                        <h5 className="font-semibold">Category Scores:</h5>
                        <ul className="list-disc ml-5 text-gray-700">
                            <li>Infrastructure Setup: {reportData.scoreInfra}%</li>
                            <li>Policy Governance: {reportData.scoreGovernance}%</li>
                            <li>Compliance Benchmarking: {reportData.scoreCompliance}%</li>
                            <li>Monitoring: {reportData.scoreMonitoring}%</li>
                            <li>Build/Deploy Security: {reportData.scoreBuildSecurity}%</li>
                        </ul>
                    </div>

                    <div className="mt-4">
                        <h5 className="font-semibold">Actions Required:</h5>
                        <ul className="list-disc ml-5 text-gray-700">
                            <li>Required Changes: {reportData.itemsRequired?.length || 0}</li>
                            <li>Recommended Changes: {reportData.itemsRecommended?.length || 0}</li>
                            <li>Advisory Actions: {reportData.itemsAdvisory?.length || 0}</li>
                        </ul>
                    </div>

                    {reportData.itemsRequired?.length > 0 && (
                        <div className="mt-4">
                            <h5 className="font-semibold text-red-700">Critical Items:</h5>
                            <ul className="list-disc ml-5 text-gray-700">
                                {reportData.itemsRequired.slice(0, 3).map((item, idx) => (
                                    <li key={idx}>{item}</li>
                                ))}
                                {reportData.itemsRequired.length > 3 && (
                                    <li>... and {reportData.itemsRequired.length - 3} more</li>
                                )}
                            </ul>
                        </div>
                    )}
                </div>
            </div>

            <button
                onClick={generateExecutiveSummary}
                disabled={generating}
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors duration-300"
            >
                {generating ? 'Generating...' : 'Download as ADOC'}
            </button>
        </div>
    );
};

export default ExecutiveSummaryTab;