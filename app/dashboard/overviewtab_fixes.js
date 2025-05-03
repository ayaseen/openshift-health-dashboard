// Fix for CircularProgress component in OverviewTab.js
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

// Fix for CategoryBar component in OverviewTab.js
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
