Contributing to OpenEBS/SCOPE

Before contributing make sure to go through all the documentation about the project.
      http://openebs.readthedocs.io/en/latest/index.html
OpenEBS/Scope is innovation in OpenSource. You are welcome to contribute in any way you can and all the help you can provide is very much appreciated.
Raise Issues on either the functionality or documentation 
Submit Changes to Improve Documentation 
Submit Proposals for new Features/Enhancements 
Submit Changes to the Source Code 
There are just a few simple guidelines that you need to follow before providing your hacks.
Please follow the code of conduct mentioned below.

Pull Request Process
1. Ensure any install or build dependencies are removed before the end of the layer when doing a build. 
2. Update the README.md with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters. 
3. Increase the version numbers in any examples files and the README.md to the new version that this Pull Request would represent. 
4. You may merge the Pull Request in once you have the sign-off of two other developers, or if you do not have permission to do that, you may request the second reviewer to merge it for you. 
Steps to generate Pull Request
1. Open your terminal window and make sure you have git installed and login to your git account.
2. Fork the openbs/scope project to your github repository
3. After that create a clone of the project from you forked repository bytyping the command :
 git clone https://github.com/USER-NAME/scope.git        (Forked repo)
4. After cloning ,work on the project make some neccesary changes or add new things and submit that changes to your forked project.
	git add -A (To add all changes made since clone)
	git commit -m "Commit message here" (Sign your work in the commit message. For more details refer below)
	git push https://github.com/USER-NAME/scope.git master
   5.Now click on New pull request.Now the authors of the project will review your changes and if valid then pull request will be accepted.Great!! you contributed to Scope.
Raising Issues
When Raising issues, please specify the following:
Setup details (like hyperconverged/dedicated), orchestration engine - kubernetes, docker swarm etc,. 
Scenario where the issue was seen to occur 
If the issue is with storage, include maya version, maya osh-status and maya omm-status. 
Errors and log messages that are thrown by the software 
Sign your work
We use the Developer Certificate of Origin (DCO) as a additional safeguard for the OpenEBS project. This is a well established and widely used mechanism to assure contributors have confirmed their right to license their contribution under the project's license. Please read developer-certificate-of-origin. If you can certify it, then just add a line to every git commit message:
  Signed-off-by: Random J Developer <random@developer.example.org>
Use your real name (sorry, no pseudonyms or anonymous contributions). If you set your user.name and user.email git configs, you can sign your commit automatically with git commit -s. You can also use git aliases like git config --global alias.ci 'commit -s'. Now you can commit with git ci and the commit will be signed.

Code of Conduct
Our Pledge
In the interest of fostering an open and welcoming environment, we as contributors and maintainers pledge to making participation in our project and our community a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.
Our Standards
Examples of behavior that contributes to creating a positive environment include:
Using welcoming and inclusive language 
Being respectful of differing viewpoints and experiences 
Gracefully accepting constructive criticism 
Focusing on what is best for the community 
Showing empathy towards other community members 
Examples of unacceptable behavior by participants include:
The use of sexualized language or imagery and unwelcome sexual attention or advances 
Trolling, insulting/derogatory comments, and personal or political attacks 
Public or private harassment 
Publishing others' private information, such as a physical or electronic address, without explicit permission 
Other conduct which could reasonably be considered inappropriate in a professional setting 
Our Responsibilities
Project maintainers are responsible for clarifying the standards of acceptable behavior and are expected to take appropriate and fair corrective action in response to any instances of unacceptable behavior.
Project maintainers do have the right and responsibility to remove, edit, or reject comments, commits, code, wiki edits, issues, and other contributions that are not aligned to this Code of Conduct, or to ban temporarily or permanently any contributor for other behaviors that they deem inappropriate, threatening, offensive, or harmful.
Scope
This Code of Conduct applies on both within project spaces and in public spaces when an individual is representing the project or its community. Examples of representing a project or community include using an official project e-mail address, posting via an official social media account, or acting as an appointed representative at an online or offline event. Representation of a project may be further defined and clarified by project maintainers.
Enforcement
Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting the project team at[admins email]. All the complaints will be reviewed and investigated and will result in a response that is deemed necessary and appropriate to the circumstances. The project team is obligated to maintain confidentiality with regard to the reporter of an incident. Further details of specific enforcement policies may be posted separately.
Project maintainers who do not follow or enforce the Code of Conduct in good faith may face temporary or permanent repercussions as determined by other members of the project's leadership.

