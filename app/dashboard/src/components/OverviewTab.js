import React from 'react';
import { getScoreColor, getScoreRating, getSeverityLevel } from '../utils/scoreUtils';

const CircularProgress = ({ value, size = 'large' }) => {
    // Make sure value is a number and default to 0 if not
    const percentage = parseFloat(value) || 0;
    const rating = getScoreRating(percentage);
    const color = getScoreColor(percentage);

    // Calculate circle properties
    const radius = size === 'large' ? 70 : 50;
    const stroke = size === 'large' ? 12 : 8;
    const normalizedRadius = radius - stroke / 2;
    const circumference = normalizedRadius * 2 * Math.PI;
    const strokeDashoffset = circumference - (percentage / 100) * circumference;

    return (
        <div className="relative flex items-center justify-center">
            <svg
                height={radius * 2}
                width={radius * 2}
                className="transform -rotate-90"
            >
                <circle
                    stroke="#e5e7eb"
                    fill="transparent"
                    strokeWidth={stroke}
                    r={normalizedRadius}
                    cx={radius}
                    cy={radius}
                />
                <circle
                    stroke={color}
                    fill="transparent"
                    strokeWidth={stroke}
                    strokeDasharray={circumference + ' ' + circumference}
                    style={{ strokeDashoffset }}
                    r={normalizedRadius}
                    cx={radius}
                    cy={radius}
                />
            </svg>
            <div className="absolute flex flex-col items-center justify-center">
                <span className={`font-bold ${size === 'large' ? 'text-3xl' : 'text-xl'}`}>
                    {Math.round(percentage)}%
                </span>
                <span className={`text-gray-600 ${size === 'large' ? 'text-lg' : 'text-sm'}`}>
                    {rating}
                </span>
            </div>
        </div>
    );
};

const CategoryBar = ({ name, score }) => {
    // Make sure score is a number and default to 0 if not
    const safeScore = parseInt(score, 10) || 0;
    const color = getScoreColor(safeScore);

    return (
        <div className="mb-6">
            <div className="flex justify-between mb-1">
                <span className="text-gray-800">{name}</span>
                <span className="font-medium">{safeScore}%</span>
            </div>
            <div className="h-3 bg-gray-200 rounded-full">
                <div
                    className="h-3 rounded-full"
                    style={{ width: `${safeScore}%`, backgroundColor: color }}
                />
            </div>
        </div>
    );
};

const ActionBox = ({ count, type, color, bgColor }) => {
    return (
        <div className={`p-4 rounded-lg border text-center ${bgColor}`}>
            <div className={`text-3xl font-bold ${color}`}>{count}</div>
            <div className={`text-sm ${color}`}>{type}</div>
        </div>
    );
};

const OverviewTab = ({ reportData }) => {
    if (!reportData) return null;

    // Get counts for different types of items
    const requiredCount = reportData.itemsRequired?.length || 0;
    const recommendedCount = reportData.itemsRecommended?.length || 0;
    const advisoryCount = reportData.itemsAdvisory?.length || 0;
    const noChangeCount = reportData.noChangeCount || 0;
    const notApplicableCount = reportData.notApplicableCount || 0;

    // Get severity style for each type
    const requiredStyle = getSeverityLevel('required');
    const recommendedStyle = getSeverityLevel('recommended');
    const advisoryStyle = getSeverityLevel('advisory');
    const noChangeStyle = getSeverityLevel('no change');
    const notApplicableStyle = getSeverityLevel('not applicable');

    return (
        <div className="space-y-6">
            {/* Overall Health Section */}
            <div className="bg-gray-50 p-6 rounded-lg shadow-sm">
                <h2 className="text-xl font-semibold mb-4">Overall Health</h2>
                <div className="flex justify-center">
                    <CircularProgress value={reportData.overallScore} />
                </div>
            </div>

            {/* Category Scores Section */}
            <div className="bg-gray-50 p-6 rounded-lg shadow-sm">
                <h2 className="text-xl font-semibold mb-6">Category Scores</h2>
                <div className="space-y-4">
                    <CategoryBar name="Infrastructure Setup" score={reportData.scoreInfra} />
                    <CategoryBar name="Policy Governance" score={reportData.scoreGovernance} />
                    <CategoryBar name="Compliance Benchmarking" score={reportData.scoreCompliance} />
                    <CategoryBar name="Monitoring" score={reportData.scoreMonitoring} />
                    <CategoryBar name="Build/Deploy Security" score={reportData.scoreBuildSecurity} />
                </div>
            </div>

            {/* Actions Required Section */}
            <div className="bg-gray-50 p-6 rounded-lg shadow-sm">
                <h2 className="text-xl font-semibold mb-6">Actions Required</h2>
                <div className="grid grid-cols-5 gap-4">
                    <ActionBox
                        count={requiredCount}
                        type="Required"
                        color={requiredStyle.color}
                        bgColor={`${requiredStyle.bgColor} ${requiredStyle.borderColor}`}
                    />
                    <ActionBox
                        count={recommendedCount}
                        type="Recommended"
                        color={recommendedStyle.color}
                        bgColor={`${recommendedStyle.bgColor} ${recommendedStyle.borderColor}`}
                    />
                    <ActionBox
                        count={advisoryCount}
                        type="Advisory"
                        color={advisoryStyle.color}
                        bgColor={`${advisoryStyle.bgColor} ${advisoryStyle.borderColor}`}
                    />
                    <ActionBox
                        count={noChangeCount}
                        type="No Change"
                        color={noChangeStyle.color}
                        bgColor={`${noChangeStyle.bgColor} ${noChangeStyle.borderColor}`}
                    />
                    <ActionBox
                        count={notApplicableCount}
                        type="Not Applicable"
                        color={notApplicableStyle.color}
                        bgColor={`${notApplicableStyle.bgColor} ${notApplicableStyle.borderColor}`}
                    />
                </div>
            </div>

            {/* Summary Statistics Section */}
            <div className="bg-gray-50 p-6 rounded-lg shadow-sm">
                <h2 className="text-xl font-semibold mb-4">Summary Statistics</h2>
                <ul className="space-y-2 text-gray-700">
                    <li><span className="font-medium">Total Items:</span> {requiredCount + recommendedCount + advisoryCount + noChangeCount + notApplicableCount}</li>
                    <li><span className="font-medium">Issues Requiring Attention:</span> {requiredCount + recommendedCount}</li>
                    <li><span className="font-medium">Issue Rate:</span> {Math.round((requiredCount + recommendedCount) / (requiredCount + recommendedCount + advisoryCount + noChangeCount) * 100)}%</li>
                    <li><span className="font-medium">Critical Issue Rate:</span> {requiredCount > 0 ? Math.round(requiredCount / (requiredCount + recommendedCount + advisoryCount + noChangeCount) * 100) : 0}%</li>
                </ul>
            </div>
        </div>
    );
};

export default OverviewTab;