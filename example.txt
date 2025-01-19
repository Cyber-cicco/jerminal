pipeline {
    agent any

    environment {
        GIT_REPO = 'git@github.com:your-organization/your-repo.git'
        DEPLOY_SERVER = 'your-deployment-server'
        DEPLOY_USER = 'deploy-user'
        SSH_CREDENTIALS_ID = 'ssh-credentials-id'
        BRANCH_NAME = "${params.BRANCH ?: 'main'}"
    }

    parameters {
        string(name: 'BRANCH', defaultValue: 'main', description: 'Git branch to build')
        booleanParam(name: 'RUN_TESTS', defaultValue: true, description: 'Run tests before deploying')
    }

    stages {
        stage('Clone Repository') {
            steps {
                sshagent(['ssh-credentials-id']) {
                    sh '''
                        echo "Cloning branch ${BRANCH_NAME}..."
                        git clone --branch ${BRANCH_NAME} ${GIT_REPO} repo
                    '''
                }
            }
        }

        stage('Build') {
            steps {
                dir('repo') {
                    sh './gradlew build'
                }
            }
        }

        stage('Test') {
            when {
                expression { params.RUN_TESTS }
            }
            steps {
                dir('repo') {
                    sh './gradlew test'
                }
            }
        }

        stage('Deploy') {
            steps {
                sshagent([SSH_CREDENTIALS_ID]) {
                    sh '''
                        echo "Deploying to ${DEPLOY_SERVER}..."
                        scp -r repo/build/libs/*.jar ${DEPLOY_USER}@${DEPLOY_SERVER}:/path/to/deploy/
                        ssh ${DEPLOY_USER}@${DEPLOY_SERVER} "systemctl restart your-app.service"
                    '''
                }
            }
        }
    }

    post {
        success {
            echo 'Pipeline completed successfully.'
        }
        failure {
            echo 'Pipeline failed.'
        }
        always {
            cleanWs() // Clean up the workspace
        }
    }
}

