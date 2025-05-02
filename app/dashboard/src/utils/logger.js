// dashboard-temp/src/utils/logger.js

/**
 * Enhanced logging utility for the dashboard
 */
class Logger {
    // Log levels
    static LEVELS = {
        DEBUG: 0,
        INFO: 1,
        WARN: 2,
        ERROR: 3
    };

    // Current log level (can be changed at runtime)
    static currentLevel = Logger.LEVELS.INFO;

    // Set log level
    static setLevel(level) {
        if (typeof level === 'string') {
            level = Logger.LEVELS[level.toUpperCase()] || Logger.LEVELS.INFO;
        }
        Logger.currentLevel = level;
        Logger.info(`Log level set to ${Object.keys(Logger.LEVELS).find(key => Logger.LEVELS[key] === level)}`);
    }

    // Debug log
    static debug(message, ...args) {
        if (Logger.currentLevel <= Logger.LEVELS.DEBUG) {
            console.debug(`[DEBUG] ${message}`, ...args);
        }
    }

    // Info log
    static info(message, ...args) {
        if (Logger.currentLevel <= Logger.LEVELS.INFO) {
            console.info(`[INFO] ${message}`, ...args);
        }
    }

    // Warning log
    static warn(message, ...args) {
        if (Logger.currentLevel <= Logger.LEVELS.WARN) {
            console.warn(`[WARN] ${message}`, ...args);
        }
    }

    // Error log
    static error(message, ...args) {
        if (Logger.currentLevel <= Logger.LEVELS.ERROR) {
            console.error(`[ERROR] ${message}`, ...args);
        }
    }

    // API request log
    static apiRequest(method, url) {
        Logger.debug(`API Request: ${method} ${url}`);
    }

    // API response log
    static apiResponse(method, url, status, timing) {
        const level = status >= 400 ? 'error' : status >= 300 ? 'warn' : 'debug';
        Logger[level](`API Response: ${method} ${url} - ${status} (${timing}ms)`);
    }

    // API error log
    static apiError(method, url, error) {
        Logger.error(`API Error: ${method} ${url}`, error);
    }

    // Create console group
    static group(label) {
        if (Logger.currentLevel <= Logger.LEVELS.DEBUG) {
            console.group(`[GROUP] ${label}`);
        }
    }

    // End console group
    static groupEnd() {
        if (Logger.currentLevel <= Logger.LEVELS.DEBUG) {
            console.groupEnd();
        }
    }

    // Log object as table
    static table(data, columns) {
        if (Logger.currentLevel <= Logger.LEVELS.DEBUG) {
            if (columns) {
                console.table(data, columns);
            } else {
                console.table(data);
            }
        }
    }

    // Log timing for performance analysis
    static time(label) {
        if (Logger.currentLevel <= Logger.LEVELS.DEBUG) {
            console.time(label);
        }
    }

    // End timing measurement
    static timeEnd(label) {
        if (Logger.currentLevel <= Logger.LEVELS.DEBUG) {
            console.timeEnd(label);
        }
    }

    // Log component lifecycle events
    static lifecycle(component, event) {
        Logger.debug(`Component: ${component} - ${event}`);
    }

    // Log data transformation
    static transform(input, output, transformName) {
        if (Logger.currentLevel <= Logger.LEVELS.DEBUG) {
            Logger.group(`Transform: ${transformName}`);
            Logger.debug('Input:', input);
            Logger.debug('Output:', output);
            Logger.groupEnd();
        }
    }

    // Log network state
    static network(online) {
        const status = online ? 'ONLINE' : 'OFFLINE';
        Logger.info(`Network Status: ${status}`);
    }

    // Log Redux/state actions
    static action(type, payload) {
        Logger.debug(`Action: ${type}`, payload);
    }

    // Log state changes
    static state(name, oldValue, newValue) {
        if (Logger.currentLevel <= Logger.LEVELS.DEBUG) {
            Logger.group(`State Change: ${name}`);
            Logger.debug('Old Value:', oldValue);
            Logger.debug('New Value:', newValue);
            Logger.groupEnd();
        }
    }
}

// Enable verbose logging in development
if (process.env.NODE_ENV !== 'production') {
    Logger.setLevel('DEBUG');
}

export default Logger;