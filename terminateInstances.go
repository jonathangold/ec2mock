// Mock EC2 Client
// Copyright (C) 2017+ Jonathan Gold and the project contributors
// Written by Jonathan Gold <info@jonathangold.ca> and the project
// contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ec2mock

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (c *MockEc2Client) TerminateInstances(in *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	for _, instanceID := range in.InstanceIds {
		for _, res := range c.Reservations {
			for _, inst := range res.Instances {
				if inst.InstanceId != instanceID {
					continue
				}
				switch aws.StringValue(inst.State.Name) {
				case ec2.InstanceStateNameRunning,
					ec2.InstanceStateNameStopped,
					ec2.InstanceStateNameStopping,
					ec2.InstanceStateNamePending:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameShuttingDown)
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameTerminated)
					}()
				case ec2.InstanceStateNameShuttingDown:
					go func() {
						time.Sleep(time.Second * StateChangeDelay)
						inst.State.Name = aws.String(ec2.InstanceStateNameTerminated)
					}()
				default:
					return nil, awserr.New(AwsErrIncorrectInstanceState, "", nil)
				}
				return nil, nil
			}
		}
	}
	return nil, nil
}
